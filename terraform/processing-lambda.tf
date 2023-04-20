data "archive_file" "processing_lambda_function_zip" {
  source_file = "../processing-lambda/main"

  type             = "zip"
  output_file_mode = "0666"
  output_path      = "../processing-lambda/main.zip"
}

resource "aws_lambda_function" "webhooks_processing_lambda" {
  function_name = "webhooks-processing-lambda"

  filename         = data.archive_file.processing_lambda_function_zip.output_path
  source_code_hash = data.archive_file.processing_lambda_function_zip.output_base64sha256

  handler     = "main"
  role        = aws_iam_role.processing_lambda_role.arn
  runtime     = "go1.x"
  memory_size = 128
  timeout     = 20 // 20 Seconds

  tracing_config {
    mode = "Active"
  }

  tags = {
    Reason = "moss-webhooks-testing"
  }
}

resource "aws_lambda_event_source_mapping" "processing_lambda_event_source_mapping" {
  event_source_arn = aws_sqs_queue.webhooks_sqs.arn
  function_name    = aws_lambda_function.webhooks_processing_lambda.arn
}