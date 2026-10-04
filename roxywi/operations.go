package roxywi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Older synchronous API responses do not contain tasks_ids.
func (c *Client) waitForTasks(ctx context.Context, response []byte) error {
	return c.waitForTasksInterval(ctx, response, 2*time.Second)
}

func (c *Client) waitForTasksInterval(ctx context.Context, response []byte, interval time.Duration) error {
	if len(response) == 0 {
		return nil
	}
	var receipt struct {
		Status string `json:"status"`
		IDs    []int  `json:"tasks_ids"`
	}
	if err := json.Unmarshal(response, &receipt); err != nil {
		return fmt.Errorf("decode operation receipt: %w", err)
	}
	if receipt.Status == "accepted" && len(receipt.IDs) == 0 {
		return fmt.Errorf("accepted operation did not return task IDs")
	}
	pending := make(map[int]bool, len(receipt.IDs))
	for _, id := range receipt.IDs {
		if id <= 0 {
			return fmt.Errorf("operation receipt contains an invalid task ID")
		}
		pending[id] = true
	}
	for len(pending) > 0 {
		for id := range pending {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("waiting for operation %d: %w; check Operations before retrying", id, err)
			}
			body, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("api/operations/%d", id), nil)
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
				// Roxy-WI 9.1 initially exposed task status through the web route only.
				body, err = c.doRequest(ctx, http.MethodGet, fmt.Sprintf("install/task-status/%d", id), nil)
			}
			if err != nil {
				return fmt.Errorf("read operation %d: %w", id, err)
			}
			var task struct {
				ID     int    `json:"task_id"`
				Status string `json:"status"`
			}
			if err := json.Unmarshal(body, &task); err != nil {
				return fmt.Errorf("decode operation %d: %w", id, err)
			}
			if task.ID != id {
				return fmt.Errorf("operation %d returned a different or missing task ID", id)
			}
			switch task.Status {
			case "completed":
				delete(pending, id)
			case "created", "published", "running":
			case "failed":
				// Worker output may contain credentials; keep it out of Terraform diagnostics.
				return fmt.Errorf("operation %d failed; check its details in Roxy-WI Operations", id)
			default:
				return fmt.Errorf("operation %d returned an unknown status", id)
			}
		}
		if len(pending) == 0 {
			return nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("waiting for operations: %w; check Operations before retrying", ctx.Err())
		case <-timer.C:
		}
	}
	return nil
}
