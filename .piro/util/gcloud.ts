import { exec } from './shell';
import { sleep } from './util';
import { getGlobalPiroInstance } from './piro';

export async function deleteExternalIp(phase: string, name: string, region = "us-west2") {
    const piro = getGlobalPiroInstance()

    const ip = getExternalIp(name)
    piro.log(phase, `address describe returned: ${ip}`)
    if (ip.indexOf("ERROR:") != -1 || ip == "") {
        piro.log(phase, `no external static IP with matching name ${name} found`)
        return
    }

    piro.log(phase, `found external static IP with matching name ${name}, will delete it`)
    const cmd = `gcloud compute addresses delete ${name} --region ${region} --quiet`
    let attempt = 0;
    for (attempt = 0; attempt < 10; attempt++) {
        let result = exec(cmd);
        if (result.code === 0 && result.stdout.indexOf("Error") == -1) {
            piro.log(phase, `external ip with name ${name} and ip ${ip} deleted`);
            break;
        } else {
            piro.log(phase, `external ip with name ${name} and ip ${ip} could not be deleted, will reattempt`)
        }
        await sleep(5000)
    }
    if (attempt == 10) {
        piro.log(phase, `could not delete the external ip with name ${name} and ip ${ip}`)
    }
}

function getExternalIp(name: string, region = "us-west2") {
    return exec(`gcloud compute addresses describe ${name} --region ${region}| grep 'address:' | cut -c 10-`, { silent: true }).trim();
}