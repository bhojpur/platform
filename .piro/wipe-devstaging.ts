import { Piro } from './util/piro'
import { wipePreviewEnvironmentAndNamespace, listAllPreviewNamespaces, helmInstallName } from './util/kubectl';
import * as fs from 'fs';
import { deleteExternalIp } from './util/gcloud';
import * as Tracing from './observability/tracing'
import { SpanStatusCode } from '@opentelemetry/api';
import { ExecOptions } from './util/shell';
import { env } from './util/util';

// Will be set once tracing has been initialized
let piro: Piro

async function wipePreviewCluster(shellOpts: ExecOptions) {
    const namespace_raw = process.env.NAMESPACE;
    const namespaces: string[] = [];
    if (namespace_raw === "<no value>" || !namespace_raw) {
        piro.log('wipe', "Going to wipe all namespaces");
        listAllPreviewNamespaces(shellOpts)
            .map(ns => namespaces.push(ns));
    } else {
        piro.log('wipe', `Going to wipe namespace ${namespace_raw}`);
        namespaces.push(namespace_raw);
    }

    for (const namespace of namespaces) {
        await wipePreviewEnvironmentAndNamespace(helmInstallName, namespace, { ...shellOpts, slice: 'wipe' });
    }
}

// clean up the dev cluster in gitpod-core-dev
async function devCleanup() {
    await wipePreviewCluster(env(""))
}

// sweeper runs in the dev cluster so we need to delete the k3s cluster first and then delete self contained namespace

Tracing.initialize()
    .then(() => {
        piro = new Piro("wipe-devstaging")
        piro.phase('wipe')
    })
    .then(() => devCleanup())
    .then(() => piro.done('wipe'))
    .then(() => piro.endAllSpans())
    .catch((err) => {
        piro.rootSpan.setStatus({
            code: SpanStatusCode.ERROR,
            message: err
        })
        piro.endAllSpans()
        console.log('Error', err)
        // Explicitly not using process.exit as we need to flush tracing, see tracing.js
        process.exitCode = 1
    });