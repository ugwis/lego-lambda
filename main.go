package main

import (
  "bytes"
  "context"
  "crypto"
  "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
  "log"
  "net/http"
  "os"
  "io/ioutil"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/registration"
  "github.com/go-acme/lego/v4/lego"
  "github.com/go-acme/lego/v4/providers/dns"

  "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-lambda-go/cfn"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3"
)

type MyEvent struct {
        Name string `json:"name"`
}

// You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}

func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func put_private(certificates *certificate.Resource, region string, bucket string, keyname string){
  body := certificates.PrivateKey

  sess := session.Must(session.NewSession())
  svc := s3.New(sess, &aws.Config{Region: aws.String(region)})
  params := &s3.PutObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(keyname),
    ACL:    aws.String("bucket-owner-full-control"),
    Body:   bytes.NewReader(body),
    ContentLength: aws.Int64(int64(len(body))),
  }
  resp, err := svc.PutObject(params)

  if err != nil {
    log.Println(err.Error())
    return
  }
  log.Println(resp)
}

func put_public(certificates *certificate.Resource, region string, bucket string, keyname string){
  url := certificates.CertURL

  resp, _ := http.Get(url)
  defer resp.Body.Close()

  body, _ := ioutil.ReadAll(resp.Body)

  sess := session.Must(session.NewSession())
  svc := s3.New(sess, &aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))})
  params := &s3.PutObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(keyname),
    ACL:    aws.String("bucket-owner-full-control"),
    Body:   bytes.NewReader(body),
    ContentLength: aws.Int64(int64(len(body))),
  }
  res, err := svc.PutObject(params)

  if err != nil {
    log.Println(err.Error())
    return
  }

  log.Println(res)
}

func handler() (string, error) {
  privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
    log.Println(err.Error())
    return "", err
	}

  email := os.Getenv("ACME_EMAIL")
  domain := os.Getenv("ACME_DOMAIN")
  region := os.Getenv("AWS_REGION")
  bucket := os.Getenv("S3_BUCKET")
  privkey := os.Getenv("S3_PRIVKEY")
  pubkey := os.Getenv("S3_PUBKEY")

	myUser := MyUser{
    Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
    log.Println(err.Error())
    return "", err
	}

  provider, err := dns.NewDNSChallengeProviderByName("route53")
  if err != nil {
    log.Println(err.Error())
    return "", err
  }
	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
    log.Println(err.Error())
    return "", err
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
    log.Println(err.Error())
    return "", err
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
    log.Println(err.Error())
    return "", err
	}
	log.Printf("%#v\n", certificates)

  put_private(certificates, region, bucket, privkey)

  put_public(certificates, region, bucket, pubkey)

  return "OK", nil
}

func lambda_handler(ctx context.Context, name MyEvent) (string, error) {
  msg, err := handler()
  return msg, err
}

func cfn_handler(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
  msg, err := handler()
  data := map[string]interface{}{}
  return msg, data, err
}

func main() {
  run_type := os.Getenv("RUN_TYPE")
  if run_type == "CloudFormation" {
    lambda.Start(cfn.LambdaWrap(cfn_handler))
  }
  if run_type == "Lambda" {
    lambda.Start(lambda_handler)
  }
  if run_type == "Shell" {
    handler()
  }
  return
}
