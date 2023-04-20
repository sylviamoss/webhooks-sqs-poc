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
