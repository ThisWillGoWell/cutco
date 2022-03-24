import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { CodePipeline, CodePipelineSource, ShellStep } from 'aws-cdk-lib/pipelines';

export class MyPipelineStack extends cdk.Stack {
  constructor(scope: Construct, id: string, branch: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const pipeline = new CodePipeline(this, 'cutcto-staging-pipeline', {
      pipelineName: 'cutco-staging-pipeline',
      synth: new ShellStep('Synth', {
        input: CodePipelineSource.gitHub('ThisWillGoWell/cutco', branch),
        commands: ['cd infra', 'npm ci', 'npm run build', 'npx cdk synth']
      })
    });
  }
}