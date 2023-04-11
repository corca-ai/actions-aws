# Self-hosted Runner Manager

## Manage your self-hosted runner

<br>

For non-ephemeral runners, there may be a lot of extra charges if you keep instances running on AWS.

With Runner Manager, you can reduce charging by stopping the runner when jobs are not running.

- Runner manager starts the runner when a job is queued.
- Runner manager stops the runner if the runner is idle for a period of time. You can set this time period yourself.

<br>

Currently, Runner Manager is valid for runners running on AWS EC2 instances.

<br>

## How to execute Runner Manager

.env

```
AWS_ACCESS_KEY          # For control runner instance
AWS_SECRET_KEY_ACCESS   # For control runner instance
AWS_CLIENT_ID           # For control runner instance
AWS_REGION              # For control runner instance
AWS_EC2_INSTANCE_ID     # For control runner instance
GITHUB_SECRET           # For webhooks
MAX_RUNNER_IDLE_TIME    # Custom time period
PORT                    # Running port for runner manager
```

<br>

1. Set up .env
2. Run with docker command

   `docker run -d -p {port} --env-file .env {image}`

3. Set up webhooks for "queued" event

   Add webhooks in GitHub repository settings

   - `Payload URL` - {ip}:{port} where your runner manager runs on
   - `Content type` - application/json
   - `Secret` - same value as GITHUB_SECRET that you put in .env
   - Select Send me everything
