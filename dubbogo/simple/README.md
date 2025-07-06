# Pixiu Dubbo-go Quick Start Sample

This document will guide you on how to run a simple service sample based on the Pixiu gateway and a Dubbo-go backend.

-----

### 1\. Steps to Run

#### Step 1: Navigate to the Sample Directory

First, clone the project and enter the directory for this sample.

```bash
# Assuming you are in the project's root directory
cd samples/dubbogo/simple/
```

#### Step 2: Prepare the Environment

Execute the following script to start the necessary dependent services (like Zookeeper) and prepare the relevant configuration files.

> **Note**: Please modify the addresses in the `benchmark/pixiu/conf.yaml` file according to your actual setup.

```bash
# This command prepares the environment for the benchmark case
./start.sh prepare benchmark
```

#### Step 3: Start the Backend Dubbo Service

Start the Dubbo-go application, which acts as the service provider.

```bash
./start.sh startServer benchmark
```

#### Step 4: Start the Pixiu Gateway

Start the Pixiu gateway in a separate terminal.

```bash
./start.sh startPixiu benchmark
```

-----

### 2\. Verify the Service

You can test whether the gateway has successfully proxied the backend service in two ways.

#### Method A: Test Directly with cURL

Open a new terminal and execute the following cURL command:

```bash
curl -s -X GET 'http://127.0.0.1:8881/api/v1/test-dubbo/user/tc?age=66'
```

**Expected Response:**

You should see a JSON output similar to the one below, which proves that Pixiu has successfully called the backend service and returned the result.

```json
{
    "age": 55,
    "code": 1,
    "iD": "0001",
    "name": "tc",
    "time": "2021-08-01T18:08:41+08:00"
}
```

#### Method B: Run the Client Test Script

You can also use the pre-configured test script to make the call.

```bash
./start.sh startTest benchmark
```

-----

### 3\. Clean Up the Environment

After testing is complete, you can run the following command to stop all the services started in this sample.

```bash
./start.sh clean
```