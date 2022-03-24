import * as cdk from 'aws-cdk-lib';
import { Construct } from "constructs";
import { AppStack } from './infra-app';

export class AppStage extends cdk.Stage {

    constructor(scope: Construct, id: string, props?: cdk.StageProps) {
        super(scope, id, props);

        const lambdaStack = new AppStack(this, 'application');
    }
}