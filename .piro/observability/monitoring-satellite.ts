import { exec } from '../util/shell';
import { getGlobalPiroInstance } from '../util/piro';
import * as shell from 'shelljs';
import * as fs from 'fs';
import { validateIPaddress } from '../util/util';

/**
 * Monitoring satellite deployment bits
 */
 export class InstallMonitoringSatelliteParams {
    pathToKubeConfig: string
    satelliteNamespace: string
    clusterName: string
    nodeExporterPort: number
    branch: string
    previewDomain: string
}

const sliceName = 'observability';

/**
 * installMonitoringSatellite installs monitoring-satellite, while updating its dependencies to the latest commit in the branch it is running.
 */
export async function installMonitoringSatellite(params: InstallMonitoringSatelliteParams) {
    const piro = getGlobalPiroInstance()
    piro.log(sliceName, `Cloning observability repository - Branch: ${params.branch}`)
    exec(`git clone --branch ${params.branch} https://roboquat:$(cat /mnt/secrets/monitoring-satellite-preview-token/token)@github.com/bhojpur/observability.git`, {silent: true})
    let currentCommit = exec(`git rev-parse HEAD`, {silent: true}).stdout.trim()
    let pwd = exec(`pwd`, {silent: true}).stdout.trim()
    piro.log(sliceName, `Updating Bhojpur.NET Platform's mixin in monitoring-satellite's jsonnetfile.json to latest commit SHA: ${currentCommit}`);

    let jsonnetFile = JSON.parse(fs.readFileSync(`${pwd}/observability/jsonnetfile.json`, 'utf8'));
    jsonnetFile.dependencies.forEach(dep => {
        if(dep.name == 'bhojpur') {
            dep.version = currentCommit
        }
    });
    fs.writeFileSync(`${pwd}/observability/jsonnetfile.json`, JSON.stringify(jsonnetFile));
    exec(`cd observability && jb update`, {slice: sliceName})

    let jsonnetRenderCmd = `cd observability && jsonnet -c -J vendor -m monitoring-satellite/manifests \
    --ext-code config="{
        namespace: '${params.satelliteNamespace}',
        clusterName: '${params.satelliteNamespace}',
        tracing: {
            honeycombAPIKey: '${process.env.HONEYCOMB_API_KEY}',
            honeycombDataset: 'preview-environments',
        },
        previewEnvironment: {
            domain: '${params.previewDomain}',
            nodeExporterPort: ${params.nodeExporterPort},
            prometheusDNS: 'prometheus-${params.previewDomain}',
            grafanaDNS: 'grafana-${params.previewDomain}',
        },
        nodeAffinity: {
            nodeSelector: {
                'bhojpur.net/workload_services': 'true',
            },
        },
    }" \
    monitoring-satellite/manifests/yaml-generator.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {} && \
    find monitoring-satellite/manifests -type f ! -name '*.yaml' ! -name '*.jsonnet'  -delete`

    piro.log(sliceName, 'rendering YAML files')
    exec(jsonnetRenderCmd, {silent: true})
    // The correct kubectl context should already be configured prior to this step
    ensureCorrectInstallationOrder()
    ensureIngressesReadiness(params)
}

async function ensureCorrectInstallationOrder(){
    const piro = getGlobalPiroInstance()

    piro.log(sliceName, 'installing monitoring-satellite')
    exec('cd observability && hack/deploy-satellite.sh', {slice: sliceName})

    deployBhojpurServiceMonitors()
    checkReadiness()
}

async function checkReadiness() {
    // For some reason prometheus' statefulset always take quite some time to get created
    // Therefore we wait a couple of seconds
    exec('sleep 30 && kubectl rollout status statefulset prometheus-k8s', {slice: sliceName})
    exec('kubectl rollout status deployment grafana', {slice: sliceName})
    exec('kubectl rollout status deployment kube-state-metrics', {slice: sliceName})
    exec('kubectl rollout status deployment otel-collector', {slice: sliceName})
    exec('kubectl rollout status daemonset node-exporter', {slice: sliceName})
}

async function deployBhojpurServiceMonitors() {
    const piro = getGlobalPiroInstance()

    piro.log(sliceName, 'installing bhojpur ServiceMonitor resources')
    exec('kubectl apply -f observability/monitoring-satellite/manifests/bhojpur/', {silent: true})
}

export function observabilityStaticChecks() {
    shell.cd('/application/operations/observability/mixins')

    if (!jsonnetFmtCheck() || !prometheusRulesCheck() || !jsonnetUnitTests()) {
        throw new Error("Observability static checks failed!")
    }
}

function jsonnetFmtCheck(): boolean {
    const piro = getGlobalPiroInstance()

    piro.log(sliceName, "Checking if jsonnet compiles and is well formated")
    let success = exec('make fmt && git diff --exit-code .', {slice: sliceName}).code == 0

    if (!success) {
        piro.fail(sliceName, "Jsonnet is badly formatted. You can fix it by running 'cd operations/observability/mixins && make fmt'");
    }

    success = exec('make lint', {slice: sliceName}).code == 0

    if (!success) {
        piro.fail(sliceName, "Jsonnet does not compile.");
    }
    return success
}

function prometheusRulesCheck(): boolean {
    const piro = getGlobalPiroInstance()

    piro.log(sliceName, "Checking if Prometheus rules are valid.")
    let success = exec("make promtool-lint", {slice: sliceName}).code == 0

    if (!success) {
        const failedMessage = `Prometheus rule validation failed. For futher reference, please read:
https://prometheus.io/docs/prometheus/latest/configuration/recording_rules/
https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/`
        piro.fail(sliceName, failedMessage)
    }
    return success
}

function jsonnetUnitTests(): boolean {
    const piro = getGlobalPiroInstance()

    piro.log(sliceName, "Running mixin unit tests")
    piro.log(sliceName, "Checking for hardcoded dashboard's datasources")

    let success = exec("make unit-tests", {slice: sliceName}).code == 0

    if (!success) {
        const failedMessage = `To make sure our dashboards work for both preview-environments and production/staging, we can't hardcode datasources. Please use datasource variables.`
        piro.fail(sliceName, failedMessage)
    }
    return success
}

function ensureIngressesReadiness(params: InstallMonitoringSatelliteParams) {
    // Read more about validating ingresses readiness
    // https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balance-ingress?hl=it#validate

    const piro = getGlobalPiroInstance()

    let grafanaIngressReady = false
    let prometheusIngressReady = false
    piro.log(sliceName, "Checking ingresses readiness")
    for(let i = 0; i < 15; i++) {
        grafanaIngressReady = ingressReady(params.satelliteNamespace, 'grafana')
        prometheusIngressReady = ingressReady(params.satelliteNamespace, 'prometheus')

        if(grafanaIngressReady && prometheusIngressReady) { break }
        piro.log(sliceName, "Trying again in 1 minute")
        exec(`sleep 60`, {slice: sliceName}) // 1 min
        i++
    }

    if (!prometheusIngressReady || !grafanaIngressReady) {
        piro.log(sliceName, "Time out while waiting for ingress readiness")
    }
}

function ingressReady(namespace: string, name: string): boolean {
    const piro = getGlobalPiroInstance()

    let ingressAddress = exec(`kubectl get ingress -n ${namespace} --no-headers ${name} | awk {'print $4'}`, { silent: true }).stdout.trim()
    if (validateIPaddress(ingressAddress)) {
        return true
    }
    piro.log(sliceName, `${name} ingress not ready.`)
    return false
}