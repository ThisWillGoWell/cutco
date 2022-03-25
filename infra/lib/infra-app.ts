import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Function, InlineCode, Runtime } from 'aws-cdk-lib/aws-lambda';
import {Cors, DomainName, LambdaIntegration, RestApi} from "aws-cdk-lib/aws-apigateway";
import {PolicyStatement} from "aws-cdk-lib/aws-iam";
import {ARecord, HostedZone, RecordSet, RecordTarget} from "aws-cdk-lib/aws-route53";
import {Certificate} from "aws-cdk-lib/aws-certificatemanager";
import * as http from "http";
import {ApiGatewayDomain, ApiGatewayv2DomainProperties} from "aws-cdk-lib/aws-route53-targets";
import {aws_route53_targets} from "aws-cdk-lib";

export class AppStack extends cdk.Stack {
    private lambdaFunction: Function
    private restAPI: RestApi

    constructor(scope: Construct, id: string, branch: string, props?: cdk.StackProps) {
        super(scope, id, props);
        function name(n: string){
            return `${props?.stackName}-${n}`
        }

        this.lambdaFunction = new Function(this, 'LambdaFunction', {
            functionName: name('gql-lambda'),
            runtime: Runtime.NODEJS_12_X,
            handler: 'index.handler',
            code: new InlineCode('exports.handler = _ => "Hello, CDK";')
        });

        let subDomain = branch

        if(subDomain == "main") {
            subDomain = "api";
        }

        this.restAPI = new RestApi(this, "RestApi", {
            restApiName: name('api-gateway'),
            domainName: {
                domainName: `${subDomain}.highground.cloud`,
                certificate:  Certificate.fromCertificateArn(this, 'cert', 'arn:aws:acm:us-east-2:571558830047:certificate/acdf6d55-a83a-4deb-b5f2-9a9e00535225')
            },
        })
        this.restAPI.root.addCorsPreflight({
            allowHeaders: ['Authorization'],
            allowOrigins: ['*'],
            maxAge: cdk.Duration.days(10),
            allowMethods: Cors.ALL_METHODS
        })

        this.restAPI.root.addResource("graph").addMethod("POST",
            new LambdaIntegration(this.lambdaFunction, {})
        )
    }
}