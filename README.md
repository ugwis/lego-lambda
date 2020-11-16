# lego-lambda
Validate with DNS01(Route 53) and generate certificates on lambda

![Architcture](https://user-images.githubusercontent.com/914815/99209252-14dee980-2806-11eb-8aae-0f4749a8f027.png "Architcture")

# Simple Usage
1. Fork this reposity
2. Set GitHub Action secrets
3. Rerun GitHub Action
4. CloudFormation Stack generate certs and put on S3 regularly

# GitHub Action Secrets

|  Name                 | Note                                        | Required  |
| --------------------- | ------------------------------------------- | --------- |
| AWS_ACCESS_KEY_ID     | AWS Access Key for deployment               | Yes       |
| AWS_SECRET_ACCESS_KEY | AWS Secret Key for deployment               | Yes       |
| AWS_REGION            | AWS Region                                  | No        |
| LAMBDA_BUCKET         | Bucket for CloudFormation Package           | No        |
| CFN_NAME              | CloudFormation Stack Name                   | No        |
| ACME_DOMAIN           | Domain to the process                       | Yes       |
| ACME_EMAIL            | Let's Encrypt User Email                    | Yes       |
| S3_BUCKET             | S3 Bucket which it put generate certs on    | Yes       |
| S3_PRIVKEY            | S3 Key which it put generate private key on | No        |
| S3_PUBKEY             | S3 Key which it put generate public key on  | No        |
