import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { CodePipeline, CodePipelineSource, ShellStep } from 'aws-cdk-lib/pipelines';
import {AppStage} from "./infra-app-pipeline";


export class PipelineStack extends cdk.Stack {
  constructor(scope: Construct, id: string, branch: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const pipeline = new CodePipeline(this,`${props?.stackName}Pipeline`, {
      pipelineName: `${props?.stackName}Pipeline`,
      synth: new ShellStep('synth', {
        input: CodePipelineSource.gitHub('ThisWillGoWell/cutco', branch),
        commands: ['cd infra', 'npm ci', 'npm run build', 'npx cdk synth'],
        primaryOutputDirectory: 'infra/cdk.out'
      })
    });

    pipeline.addStage(new AppStage(this, `${props?.stackName}AppStack`, branch,{
      env: props?.env
    }));
  }
}