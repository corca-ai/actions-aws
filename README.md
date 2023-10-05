# Self-hosted Runner Manager

### Manage your self-hosted runner

<br>

For non-ephemeral runners, there may be a lot of extra charges if you keep instances running on AWS.

With Runner Manager, you can reduce charging by stopping the runner when jobs are not running.

- Runner manager starts the runner when a job is queued.
- Runner manager stops the runner if the runner is idle for a period of time. You can set this time period yourself.

<br>

Currently, Runner Manager is valid for runners running on AWS EC2 instances.

<br>

### How to execute Runner Manager

.env

```text
# All values should be provided as they are (PORT="4000" - (X) PORT=4000 - (O))
AWS_ACCESS_KEY_ID       # For control runner instance
AWS_SECRET_ACCESS_KEY   # For control runner instance
AWS_CLIENT_ID           # For control runner instance
AWS_REGION              # For control runner instance (Default: ap-northeast-2)
AWS_EC2_INSTANCE_ID     # For control runner instance (Optional)
GITHUB_SECRET           # For webhooks
MAX_RUNNER_IDLE_TIME    # Custom idle time period (Default: 30m)
RUNNER_WAIT_TIMEOUT     # Custom waiting runner time period (Default: 3m)
PORT                    # Running port for runner manager
```

<br>

1. Set up .env
   - If `AWS_EC2_INSTANCE_ID` is not provided, the runner manager will launch a new instance. In this case, you should also provide the following additional environment variables. `GITHUB_URL` and `GITHUB_TOKEN` are used for [generating tokens](https://docs.github.com/en/rest/actions/self-hosted-runners?apiVersion=2022-11-28#create-configuration-for-a-just-in-time-runner-for-a-repository).

   ```text
   AWS_EC2_VOLUME_SIZE      # For launching runner instance (Default: 16)
   GITHUB_URL               # For launching runner instance (github repository url)
   GITHUB_TOKEN             # For launching runner instance (personal access token)
   LABELS                   # For launching runner instance (Optional, should be separated by space)
   ```

2. Run with docker command

   `docker run -d -p {image-port}:{host-port} --env-file .env {image}`

3. Set up webhooks for "queued" event

   Add webhooks in GitHub repository settings

   - `Payload URL` - {host-ip}:{host-port} where your runner manager runs on
   - `Content type` - application/json
   - `Secret` - same value as GITHUB_SECRET that you put in .env
   - Select Send me everything
