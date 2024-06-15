# Ascenda Loyalty Points System - Backend

Backend Repository for Ascenda LP System leveraging Amazon Web Services (AWS) Serverless Infrastructure.

## Project Overview

- **Background:** Ascenda operates a comprehensive administrative subsystem crucial for managing back-office operations and facilitating customer support across various departments.
- **Objective:** The project aims to develop an advanced administrative system that incorporates robust authentication, fine-grained access control, seamless integration with multiple backend applications, and comprehensive logging functionalities.

## Solution Architecture
![alt text](docs/architecture.png)
- Our architecture employs a serverless approach using AWS services to ensure scalability, resilience, and ease of maintenance. The core components include AWS Lambda for processing, API Gateway for managing API requests, and Amazon Aurora for database management. 
- This architecture is designed to abstract infrastructure management, streamline deployment processes, and reduce operational overhead.

## Technologies Utilised
<div align="center">
  <h3>Cloud Platform & Services</h3>
  <img src="https://uxwing.com/wp-content/themes/uxwing/download/brands-and-social-media/aws-icon.png" alt="AWS" width="88"/>
  <p><strong>AWS Services</strong></p>
  <img src="https://www.brcline.com/wp-content/uploads/2021/09/aws-lambda-logo.png" alt="AWS Lambda" height="60"/>
  <img src="https://www.prolim.com/wp-content/uploads/2019/09/amazon-api-gatewat-1.jpg" alt="AWS API Gateway" height="60"/>
  <img src="https://cloudkul.com/blog/wp-content/uploads/2022/03/AWS-WAF-logo.png" alt="AWS WAF" height="60"/>
  <img src="https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSxFjLuV6wjnZ3d15kwaxPqzhVs89wP5h2i_Q&s" alt="AWS Amplify" height="60"/>

  <img src="https://miro.medium.com/v2/resize:fit:600/1*w7l_juI3zKXit-dXpSS0Mg.png" alt="AWS RDS (Aurora)" height="60"/>
  <img src="https://miro.medium.com/v2/resize:fit:556/1*tTedvyOfnCu_8O26I3vlDA.png" alt="AWS DynamoDB" height="60"/>

  <h3>DevOps</h3>
  <img src="https://i0.wp.com/foxutech.com/wp-content/uploads/2017/09/AWS-CloudFormation-1.png?fit=640%2C366&ssl=1" alt="AWS CloudFormation" height="60"/>
  <img src="https://gdm-catalog-fmapi-prod.imgix.net/ProductLogo/77befea2-7041-4c6c-9ec7-d75bb60b21c6.png" alt="Terraform" width="60"/>
  <img src="https://miro.medium.com/v2/resize:fit:1075/1*5WC9rtIa0KLXfRrC8Swf1w.png" alt="GitHub Actions" height="60"/>

  <h3>Programming Languages & Framework</h3>
  <img src="https://static-00.iconduck.com/assets.00/react-original-wordmark-icon-840x1024-vhmauxp6.png" alt="React.js" height="40"/>
  <img src="https://blog.golang.org/go-brand/Go-Logo/SVG/Go-Logo_Blue.svg" alt="Golang" width="60"/>
</div>


## Setting up the backend

```sh
# Build & Run package locally
make build-run

# Build & Deploy
make deploy

# Teardown (Do NOT teardown unless necessary)
make teardown
```

## Others

- Link to Frontend Repository
- Link to Terraform
