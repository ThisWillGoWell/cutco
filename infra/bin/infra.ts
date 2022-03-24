#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { MyPipelineStack } from '../lib/infra-stack';

const app = new cdk.App();
new MyPipelineStack(app, 'MyPipelineStack', {
    env: {
        account: '111111111111',
        region: 'eu-west-1',
    }
});

app.synth();