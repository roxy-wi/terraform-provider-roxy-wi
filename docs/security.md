# Security guidance

## Terraform state

The provider marks credentials as sensitive, so Terraform redacts them in normal CLI output. Sensitive values can still be present in Terraform state because the provider currently uses Terraform Plugin SDK v2, which does not support write-only resource arguments.

Protect state as secret material:

- use a remote backend with encryption at rest and in transit;
- restrict read access to the smallest possible group;
- enable backend versioning and audit logs;
- never commit local state or plan files;
- rotate credentials if state may have been exposed.

Read operations do not copy password, passphrase, private-key, or S3 credential values returned by APIs back into state.

## Planned write-only arguments

Terraform Plugin SDK v2 supports write-only arguments when used with Terraform 1.11 or later. Enabling them in this provider is intentionally reserved for a major release because the current provider supports Terraform 1.7 and existing configurations use state-backed secret attributes.

The migration will keep resource addresses stable:

1. raise the minimum supported Terraform version to 1.11;
2. add write-only `_wo` attributes and matching version attributes to `roxywi_user`, `roxywi_ssh_cred`, and `roxywi_backup_s3`;
3. read write-only values from raw resource configuration and add compatibility/state-upgrade tests;
4. warn when legacy state-backed secret attributes are used, then remove them in the following major release.

Until that migration is complete, consumers must treat Terraform state as containing secrets.
