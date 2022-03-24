
xport CDK_NEW_BOOTSTRAP=1
npx cdk bootstrap aws://571558830047/us-east-2 \
    --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess \
    aws://571558830047/us-east-2

