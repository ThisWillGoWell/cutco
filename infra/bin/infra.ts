#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { MyPipelineStack } from '../lib/infra-pipeline';

const app = new cdk.App();
new MyPipelineStack(app, 'PipelineStack', 'staging', {
    env: {
        account: '571558830047',
        region: 'us-east-2',
    }
});

app.synth();