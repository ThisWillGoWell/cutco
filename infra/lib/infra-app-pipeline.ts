import * as cdk from 'aws-cdk-lib';
import { Construct } from "constructs";
import { AppStack } from './infra-app';

export class AppStage extends cdk.Stage {

    constructor(scope: Construct, id: string,  branch: string, props?: cdk.StageProps) {
        super(scope, id, props);
        new AppStack(this, branch,`${branch}AppStack`);
    }
}