## Conductor Sagas

This is a case study done as part of my final project for my bachelors in Information Systems at UFMG. The text report is available at [TODO: Add DCC link]. 

In summary this is a case study of how sagas can be implemented inside Conductor and which approach has the most potential.


## How to run

Dependencies:
- `docker`
- `node`
- `npm`

Running the application: 
- Add your API key to datadog-agent definition inside the `docker-compose.yaml` file if you want to collect the metrics
- Run `docker compose up` from repository root to start services
- Clone `https://github.com/danhenrik/conductor`
- Add your API key to `server/src/main/resources/application.properties` in the `management.datadog.metrics.export.apiKey` variable
- `cd` into docker and run `docker compose up`
- If you added the DataDog API key to the configurations file the metrics should show up on your DataDog account
- Enter `http://localhost:8127` and in the `definitions` tab add the workflows available in the `conductor-workflows` folder

Running the performance tests:
- `cd` into perf-tests and run `npm install`
- Run `npm install --global k6@latest`
- To generating the test data use the command `node generate-test-mass.js <Number of requests> <Number of batches>`
- Set the same number of batches you passed as parameter for the previous command as the value of the `BATCH_COUNT` constant inside `script.js` (line 7)
- Also configure the workflow you want to test (either `travel_booking_http_saga_wf`,`travel_booking_worker_saga_wf` or `travel_booking_segregated_saga_wf`) in the `WORKFLOW_NAME` constant (line 8)
- To execute the test run `k6 run script.js`