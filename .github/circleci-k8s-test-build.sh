#!/bin/bash

if [ -z "${REPO_NAME}" ] || [ -z "${CIRCLECI_DEV_API_TOKEN}" ] || [ -z "${JOB_COUNT}" ]; then
    echo "Missing CIRCLECI_DEV_API_TOKEN secret"
    exit 1
fi

echo "Running ${REPO_NAME} build on CircleCI"
echo "Project link: https://circleci.com/gh/wunderio/workflows/${REPO_NAME}"

base_api_url="https://circleci.com/api/v1.1/project/github/wunderio/${REPO_NAME}"
# Trigger a new deployment.
curl -s -X POST $base_api_url/build?circle-token=${CIRCLECI_DEV_API_TOKEN}
sleep 10
# Wait for deployment to be complete
while curl -s "$base_api_url?circle-token=${CIRCLECI_DEV_API_TOKEN}&limit=${JOB_COUNT}" | jq -e 'any(.[]; (.status == "running") or (.status == "queued"))' > /dev/null
do
echo "still running"
sleep 10
done
# Test that the build was successful
curl -s "$base_api_url?circle-token=${CIRCLECI_DEV_API_TOKEN}&limit=${JOB_COUNT}" | jq '.[] | { job_name: .workflows.job_name, status: .status }'
curl -s "$base_api_url?circle-token=${CIRCLECI_DEV_API_TOKEN}&limit=${JOB_COUNT}" | jq -e 'all(.[]; .status == "success")' > /dev/null
