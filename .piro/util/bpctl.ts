import * as shell from 'shelljs';
import { ExecOptions } from './shell';

export function buildBpctlBinary() {
    shell.exec(`cd /application/dev/bpctl && go build && cd -`)
}

export function printClustersList(shellOpts: ExecOptions): string {
    const result = shell.exec(`/application/dev/bpctl/bpctl clusters list`, { ...shellOpts, async: false }).trim()
    return result
}

export function uncordonCluster(name: string, shellOpts: ExecOptions): string {
    const result = shell.exec(`/application/dev/bpctl/bpctl clusters uncordon --name=${name}`, { ...shellOpts, async: false }).trim();
    return result
}

export function registerCluster(name: string, url: string, shellOpts: ExecOptions): string {
    const cmd = `/application/dev/bpctl/bpctl clusters register \
	--name ${name} \
	--hint-cordoned \
	--hint-govern \
	--tls-path ./wsman-tls \
	--url ${url}`;
    const result = shell.exec(cmd, { ...shellOpts, async: false }).trim();
    return result
}

export function getClusterTLS(shellOpts: ExecOptions): string {
    const result = shell.exec(`/application/dev/bpctl/bpctl clusters get-tls-config`, { ...shellOpts, async: false }).trim()
    return result
}
