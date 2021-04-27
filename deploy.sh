#!/bin/sh

set +e

gcloud functions deploy gcp-monitoring-to-teams \
    --entry-point=F \
    --memory=128MB \
    --region=us-central1 \
    --runtime=go113 \
    --env-vars-file=.env.yaml \
    --trigger-http \
    --timeout=10s
