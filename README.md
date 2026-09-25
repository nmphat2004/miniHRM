# Mini HRM

Mini HRM is a local-first department and employee management application. Run DynamoDB Local with `docker compose up -d`, start the Go API with `npm run backend:dev`, and start Next.js with `npm run frontend:dev` (or use `npm run dev` where WSL/Docker is configured).

## Architecture and DynamoDB keys

The Go API uses a single DynamoDB table (`mini_hrm_table`) with `PK` / `SK` and the `GSI1` index (`GSI1PK` / `GSI1SK`). Entity records and append-only audit entries share the table:

For AWS DynamoDB, configure `AWS_REGION`, `AWS_PROFILE`, and `DYNAMODB_TABLE_NAME` in `backend/.env` (or the process environment), then authenticate the profile with `aws sso login --profile <profile>`. When `AWS_PROFILE` is set and `DYNAMODB_ENDPOINT` is omitted, the server and seed command use the AWS SDK credential chain. To use DynamoDB Local, set `DYNAMODB_ENDPOINT=http://localhost:8000`; without an AWS profile or endpoint, local mode is the default. The seed command loads `backend/.env` before connecting.

| Record / access pattern | PK | SK | GSI1PK | GSI1SK |
| --- | --- | --- | --- | --- |
| Department metadata | `DEPT#<id>` | `METADATA` | `DEPT_PARENT#<parentId-or-ROOT>` | `PATH#<path>` |
| Department unique code | `DEPT_CODE#<code>` | `METADATA` | — | — |
| Employee metadata | `EMP#<id>` | `METADATA` | `DEPT#<departmentId>` | `STATUS_JOINED#<status>#<joinedAt>` |
| Employee unique code / email | `EMP_CODE#<code>` / `EMP_EMAIL#<email>` | `METADATA` | — | — |
| Audit event | `AUDIT#<targetId>` | `LOG#<occurredAt>#<id>` | `AUDIT_ALL` | `LOG#<occurredAt>#<id>` |
| Idempotency result | `IDEMPOTENCY#<key>` | `METADATA` | — | — |

Department listings walk the `DEPT_PARENT#...` GSI from roots or the caller's current department; employee listing queries the `DEPT#...` GSI only for departments inside that scope. No application-side filtering of a company-wide `Scan` is used for those list operations. Admin queries traverse the whole hierarchy because their authorized scope is the whole company. User scope is resolved from the current employee/department records on each request, not trusted from stale token department claims.

Creates write the new entity, uniqueness records, audit entry, and (when supplied) the idempotency response in one `TransactWriteItems`. Idempotency results expire after 24 hours (`expiresAt` TTL); the API also checks expiry itself because DynamoDB TTL deletion is asynchronous. Repeating the same create request/key returns the original response.

## Department move limits

A department move rewrites the moved department, all descendant department paths, employee department paths in the moved subtree, and its audit event in one atomic transaction. Before writing, the service counts all affected records. It rejects a move that would exceed DynamoDB's 100-item transaction limit, leaving the tree unchanged. This implementation deliberately favors atomic consistency over multi-phase/chunked moves; large subtrees must be split/restructured before moving. Transactions remain subject to DynamoDB's 4 MB aggregate transaction-size limit as well.

## Authentication

The frontend stores only display claims in local storage. Authentication tokens are delivered only in an `HttpOnly`, `SameSite=Strict` session cookie. Local development defaults to non-secure cookies on HTTP; set `COOKIE_SECURE=true` behind HTTPS. Mock auth is the default for offline development. To use Cognito, configure `COGNITO_USER_POOL_ID`, `COGNITO_CLIENT_ID`, and `AWS_REGION`; the app uses Cognito `USER_PASSWORD_AUTH` and validates RS256 ID tokens against the pool JWKS. The Cognito app client must allow that auth flow and issue `custom:role` (admin/manager/employee) claims; add `custom:employeeId` to users so data scopes resolve to the corresponding Mini HRM employee record (otherwise Cognito `sub` is used as the record ID).

## Seed accounts and test data

Run `npm run backend:seed` after DynamoDB Local is available. Seed data includes the root, two parallel branches with three or more employees apiece, two managers in different branches, an archived department, and a resigned employee. Local mock credentials are displayed on the login page.

## Employee profile photos

Admins and managers may add, replace, or remove an employee's profile photo through the profile edit form (managers remain scoped to their own branch). The browser converts JPG/PNG/WebP inputs to a small JPEG; the API validates JPEG/PNG content, dimensions at most 512×512, and a decoded size below 64 KiB. When `AVATAR_S3_BUCKET` is set, Go uploads the private image to S3 and DynamoDB stores only its object key. Reads return a one-hour presigned URL. The upload happens before the employee/audit transaction; failed DB writes trigger best-effort removal of the new object, while the previous object is deleted after a successful commit. Audit entries record only whether a photo exists, never its binary content. Without an S3 bucket configured, avatar writes are disabled; legacy data URLs remain readable for migration.

Configure `AWS_REGION`, `AVATAR_S3_BUCKET`, and the AWS SDK credential chain (local AWS profile/SSO or the deployment role) in `backend/.env` or the process environment. The server loads `backend/.env` at startup without printing its values. The role needs `s3:PutObject`, `s3:GetObject`, and `s3:DeleteObject` on `arn:aws:s3:::<bucket>/employees/*`. Keep S3 Block Public Access enabled; uploads go through the Go API, so the bucket does not need browser CORS rules.

## Frontend design workflow

The project includes the `Leonxlnx/taste-skill` design skills under `.agents/skills/`, with the installed skill list in `skills-lock.json`. For an existing-dashboard redesign, use `redesign-existing-projects` with `design-taste-frontend` as the design-system reference; preserve this app's Mini HRM slate/indigo tokens and workflows. To refresh the skills, run `npx skills add Leonxlnx/taste-skill --agent opencode -y`.
