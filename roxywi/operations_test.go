package roxywi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForTasks(t *testing.T) {
	for _, tc := range []struct {
		name, receipt, status string
		fail                  bool
	}{
		{"completed", `{"tasks_ids":[7]}`, "completed", false},
		{"failed", `{"tasks_ids":[7]}`, "failed", true},
		{"unknown", `{"tasks_ids":[7]}`, "other", true},
		{"missing status", `{"tasks_ids":[7]}`, "", true},
		{"invalid id", `{"tasks_ids":[0]}`, "completed", true},
		{"malformed ids", `{"tasks_ids":["7"]}`, "completed", true},
		{"missing ids", `{"status":"accepted"}`, "completed", true},
		{"legacy response", `{"id":7}`, "completed", false},
		{"draft", `{"status":"draft","tasks_ids":[]}`, "completed", false},
		{"empty delete", ``, "completed", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/api/operations/7" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
				writeTestJSON(t, w, map[string]interface{}{"task_id": 7, "status": tc.status, "error": "private-worker-secret"})
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := config.Client.waitForTasksInterval(ctx, []byte(tc.receipt), time.Millisecond)
			if (err != nil) != tc.fail {
				t.Fatalf("error = %v, want failure %v", err, tc.fail)
			}
			if err != nil && strings.Contains(err.Error(), "private-worker-secret") {
				t.Fatal("worker output leaked")
			}
		})
	}
}

func TestWaitForTasksPollsAllOperations(t *testing.T) {
	var first, second atomic.Int32
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		id, count := 7, first.Add(1)
		if r.URL.Path == "/api/operations/8" {
			id, count = 8, second.Add(1)
		}
		state := "running"
		if count > 2 {
			state = "completed"
		}
		writeTestJSON(t, w, map[string]interface{}{"task_id": id, "status": state})
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := config.Client.waitForTasksInterval(ctx, []byte(`{"tasks_ids":[7,8]}`), time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if second.Load() < 3 {
		t.Fatal("did not wait for all operations")
	}
}

func TestWaitForTasksCancellation(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]interface{}{"task_id": 7, "status": "published"})
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	err := config.Client.waitForTasksInterval(ctx, []byte(`{"tasks_ids":[7]}`), time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestWaitForTasksLegacyRouteAndReadErrors(t *testing.T) {
	for _, status := range []int{200, 403, 404, 500} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/operations/7" {
					http.NotFound(w, r)
					return
				}
				if r.URL.Path != "/install/task-status/7" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.WriteHeader(status)
				writeTestJSON(t, w, map[string]interface{}{"task_id": 7, "status": "completed"})
			})
			err := config.Client.waitForTasksInterval(context.Background(), []byte(`{"tasks_ids":[7]}`), time.Millisecond)
			if (err != nil) != (status != 200) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
