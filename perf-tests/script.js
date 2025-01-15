import http from "k6/http";
import { check, sleep } from "k6";
import { SharedArray } from "k6/data";
import { Counter } from "k6/metrics";
import papaparse from './node_modules/papaparse/papaparse.js'; 

const CONDUCTOR_URL = "http://localhost:8080/api/workflow";
const BATCH_COUNT = 40_000; 

export let options = {
  vus: 50,
  duration: "30s",
};

const globalIndex = new Counter("globalIndex");

const batchFiles = new SharedArray("batchFiles", function () {
  var files = new Array();
  for (let i = 0; i < BATCH_COUNT; i++) {
    files.push(`batches/batch-${i}.csv`);
  }
  return files;
});

export default function () {
  const currentIndex = globalIndex.add(1) - 1;
  if (currentIndex >= batchFiles.length) {
    process.exit(0);
  }

  const batchFile = batchFiles[currentIndex];
  console.log(batchFile);
  if (batchFile) {
    const csvData = open(batchFile);
    const parsedData = papaparse.parse(csvData, { header: true }).data;
    for (const data of parsedData) {
      let res = http.post(CONDUCTOR_URL, data);
      check(res, { "status was 200": (r) => r.status === 200 });
      sleep(1);
      logWithTimestamp(
        `Response Status: ${response.status}.\n Response body: ${response.body}`
      );
    }
  } else {
    logWithTimestamp("Fila de batches vazia!");
  }
}

function logWithTimestamp(...args) {
  console.log(new Date().toISOString(), ...args);
}
