#!/bin/bash
set -euo pipefail

aws ec2 describe-instances \
    --output text \
    --filters "Name=instance-state-name,Values=running" \
    --query 'Reservations[*].Instances[*].{InstanceId: InstanceId, State: State.Name, IP: PublicIpAddress}'
