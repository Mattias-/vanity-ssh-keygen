#!/bin/bash
set -euo pipefail

aws ec2 describe-instances \
    --output json \
    --filters "Name=instance-state-name,Values=running" \
    --query 'Reservations[].Instances[].{InstanceId: InstanceId, State: State.Name, IP: PublicIpAddress}' |
    jq -r '.[] | "- \(.IP):9100 # \(.InstanceId)"'
