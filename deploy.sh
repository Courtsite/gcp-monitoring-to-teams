#!/bin/sh

set +e

gcloud alpha functions deploy stackdriver-to-discord \
  --entry-point F \
  --memory 128MB \
  --region us-central1 \
  --runtime go111 \
  --env-vars-file .env.yaml \
  --trigger-http
