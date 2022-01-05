#!/bin/bash

set -euo pipefail

NAMESPACE=$1

if [[ -z ${NAMESPACE} ]]; then
   echo "One or more input params were invalid. The params we received were: ${NAMESPACE}"
   exit 1
fi

echo "Removing Bhojpur.NET Platform in namespace ${NAMESPACE}"
kubectl get configmap bhojpur-app -n "${NAMESPACE}" -o jsonpath='{.data.app\.yaml}' | kubectl delete --ignore-not-found=true -f -

echo "Removing Bhojpur.NET Platform storage from ${NAMESPACE}"
kubectl -n "${NAMESPACE}" delete pvc data-mysql-0
# the installer includes the minio PVC in it's config mpap, this is a "just in case"
kubectl -n "${NAMESPACE}" delete pvc minio || true

echo "Successfully removed Bhojpur.NET Platform from ${NAMESPACE}"