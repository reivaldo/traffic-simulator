#!/bin/sh
# NATS JetStream setup script for local dev
set -e

nats stream add messages --subjects "messages.*" --storage file --retention limits --max-msgs 100000 --max-bytes 1GB --max-age 72h --ack --dupe-window 2m || true
nats consumer add messages ingestor --filter "messages.incoming" --ack --deliver all --replay instant --max-deliver 10 || true
nats consumer add messages processor --filter "messages.incoming" --ack --deliver all --replay instant --max-deliver 10 || true
nats consumer add messages sender --filter "messages.processed" --ack --deliver all --replay instant --max-deliver 10 || true
