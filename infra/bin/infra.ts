#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import {PipelineStack} from "../lib/infra-pipeline";

const app = new cdk.App();
new PipelineStack(app, 'cutco-staging', 'staging', {
    env: {
        account: '571558830047',
        region: 'us-east-2',
    }
});

new PipelineStack(app, 'cutco-prod', 'main', {
    env: {
        account: '571558830047',
        region: 'us-east-2',
    }
});

app.synth();