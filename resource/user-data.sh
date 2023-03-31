#!/bin/bash
export ACTIONS_RUNNER_VERSION={{ ACTIONS_RUNNER_VERSION }}
export GITHUB_URL={{ GITHUB_URL }}
export GITHUB_TOKEN={{ GITHUB_TOKEN }}
export PUBLIC_IP=$(dig +short myip.opendns.com @resolver1.opendns.com)
export RUN_SCRIPT=/home/ubuntu/runner.sh

apt-get update && apt-get install -y curl unzip ca-certificates gnupg

echo "Installing Docker"
mkdir -m 0755 -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null
chmod a+r /etc/apt/keyrings/docker.gpg
apt-get update && apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
useradd -ms /bin/bash runner
groupadd docker
usermod -aG docker runner

echo "Installing AWS CLI"
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
./aws/install

cd /home/ubuntu

echo "Installing Actions Runner"
curl -o actions-runner-linux-x64-$ACTIONS_RUNNER_VERSION.tar.gz -L https://github.com/actions/runner/releases/download/v$ACTIONS_RUNNER_VERSION/actions-runner-linux-x64-$ACTIONS_RUNNER_VERSION.tar.gz
tar xzf ./actions-runner-linux-x64-$ACTIONS_RUNNER_VERSION.tar.gz
./bin/installdependencies.sh

cat <<- EOF > $RUN_SCRIPT
newgrp docker
./config.sh --url "$GITHUB_URL" --pat "$GITHUB_TOKEN" --name $PUBLIC_IP --runnergroup default --work _work --labels self-hosted,Linux,X64 > config.log
nohup ./runsvc.sh > action.log &
EOF

chown -R ubuntu:ubuntu /home/ubuntu

su ubuntu -c $RUN_SCRIPT

