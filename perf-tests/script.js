import http from "k6/http";
import { SharedArray } from "k6/data";
import { sleep } from 'k6';
import papaparse from "./node_modules/papaparse/papaparse.js";

const CONDUCTOR_URL = "http://localhost:8080/api/workflow";
const BATCH_COUNT = 500;
const WORKFLOW_NAME = "travel_booking_http_saga_wf";
const WORKFLOW_VERSION = 1;

export let options = {
  vus: BATCH_COUNT,
  iterations: BATCH_COUNT, // Guarantees that each VU will process once
};

const batchFiles = new SharedArray("batchFiles", function () {
  var files = new Array();
  for (let i = 0; i < BATCH_COUNT; i++) {
    var fileData = open(`batches/batch-${i}.csv`);
    files.push(fileData);
  }

  return files;
});

export default function () {
  const currentIndex  = ((__VU - 1));

  const csvData = batchFiles[currentIndex];
  if (csvData) {
    const parsedData = papaparse.parse(csvData, { header: true }).data;
    for (const data of parsedData) {
      let requestBody = {
        name: WORKFLOW_NAME,
        version: WORKFLOW_VERSION,
        input: data,
      };

      let res = http.post(
        CONDUCTOR_URL,
        JSON.stringify(requestBody, (key, value) =>
          isNaN(value) ? value : +value
        ),
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );

      logWithTimestamp(
        `Response Status: ${res.status} for index ${currentIndex}.\n ExecutionId: ${res.body}`
      );
      sleep(1)
    }
  } else {
    logWithTimestamp("Fila de batches vazia!");
  }
}

function logWithTimestamp(...args) {
  console.log(new Date().toISOString(), ...args);
}