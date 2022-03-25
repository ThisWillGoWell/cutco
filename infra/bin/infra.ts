#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import {PipelineStack} from "../lib/infra-pipeline";

const app = new cdk.App();
new PipelineStack(app, 'CutcoStaging', 'staging', {
    env: {
        account: '571558830047',
        region: 'us-east-2',
    }
});

new PipelineStack(app, 'CutcoProd', 'main', {
    env: {
        account: '571558830047',
        region: 'us-east-2',
    }
});

app.synth();