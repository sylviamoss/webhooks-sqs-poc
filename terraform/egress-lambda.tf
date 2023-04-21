data "archive_file" "egress_lambda_function_zip" {
  source_file = "../egress-lambda/main"

  type             = "zip"
  output_file_mode = "0666"
  output_path      = "../egress-lambda/main.zip"
}

resource "aws_lambda_function" "webhooks_egress_lambda" {
  function_name = "webhooks-egress-lambda"

  filename         = data.archive_file.egress_lambda_function_zip.output_path
  source_code_hash = data.archive_file.egress_lambda_function_zip.output_base64sha256

  handler     = "main"
  role        = aws_iam_role.egress_lambda_role.arn
  runtime     = "go1.x"
  memory_size = 128
  timeout     = 20 // 20 Seconds

  tracing_config {
    mode = "Active"
  }
}

resource "aws_iam_role" "egress_lambda_role" {
  name = "egress-lambda-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      },
    ]
  })
  tags = {
    Reason = "moss-webhooks-testing"
  }
}

resource "aws_iam_role_policy_attachment" "egress_lambda_basic_role_policy" {
  role       = aws_iam_role.egress_lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
