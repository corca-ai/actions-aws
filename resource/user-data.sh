#!/bin/bash
export ACTIONS_RUNNER_VERSION={{ ACTIONS_RUNNER_VERSION }}
export GITHUB_URL={{ GITHUB_URL }}
export GITHUB_TOKEN={{ GITHUB_TOKEN }}

apt-get update && apt-get install -y curl unzip
useradd -ms /bin/bash runner
groupadd docker
usermod -aG docker runner

mkdir /home/runner
cd /home/runner

curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install

curl -o actions-runner-linux-x64-$ACTIONS_RUNNER_VERSION.tar.gz -L https://github.com/actions/runner/releases/download/v$ACTIONS_RUNNER_VERSION/actions-runner-linux-x64-$ACTIONS_RUNNER_VERSION.tar.gz
tar xzf ./actions-runner-linux-x64-$ACTIONS_RUNNER_VERSION.tar.gz
./bin/installdependencies.sh

chown -R runner:runner /home/runner

su - runner

newgrp docker

./config.sh --ephemeral --url "$GITHUB_URL" --token "$GITHUB_TOKEN"
./run.sh

echo "sudo halt" | at now + 3 minutes
