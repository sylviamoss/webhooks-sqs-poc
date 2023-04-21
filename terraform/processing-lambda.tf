#data "archive_file" "processing_lambda_function_zip" {
#  source_file = "../processing-lambda/main"
#
#  type             = "zip"
#  output_file_mode = "0666"
#  output_path      = "../processing-lambda/main.zip"
#}
#
#resource "aws_lambda_function" "webhooks_processing_lambda" {
#  function_name = "webhooks-processing-lambda"
#
#  filename         = data.archive_file.processing_lambda_function_zip.output_path
#  source_code_hash = data.archive_file.processing_lambda_function_zip.output_base64sha256
#
#  handler     = "main"
#  role        = aws_iam_role.processing_lambda_role.arn
#  runtime     = "go1.x"
#  memory_size = 128
#  timeout     = 20 // 20 Seconds
#
#  tracing_config {
#    mode = "Active"
#  }
#
#  tags = {
#    Reason = "moss-webhooks-testing"
#  }
#}
#
#resource "aws_lambda_event_source_mapping" "processing_lambda_event_source_mapping" {
#  event_source_arn = aws_sqs_queue.webhooks_sqs.arn
#  function_name    = aws_lambda_function.webhooks_processing_lambda.arn
#}

#data "aws_iam_policy_document" "processing_lambda_policy" {
#  // Allow processing lambda to invoke egress lambda.
#  statement {
#    effect = "Allow"
#    actions = [
#      "lambda:InvokeFunction",
#    ]
#    resources = [aws_lambda_function.webhooks_egress_lambda.arn]
#  }
#  statement {
#    effect = "Allow"
#    actions = [
#      "sqs:GetQueueUrl",
#      "sqs:ChangeMessageVisibility",
#    ]
#    resources = [aws_sqs_queue.webhooks_sqs.arn]
#  }
#}
#
#resource "aws_iam_role" "processing_lambda_role" {
#  name = "processing-lambda-role"
#  assume_role_policy = jsonencode({
#    Version = "2012-10-17"
#    Statement = [
#      {
#        Action = "sts:AssumeRole"
#        Effect = "Allow"
#        Sid    = ""
#        Principal = {
#          Service = "lambda.amazonaws.com"
#        }
#      },
#    ]
#  })
#  tags = {
#    Reason = "moss-webhooks-testing"
#  }
#}
#
#resource "aws_iam_role_policy" "processing_lambda_role_policy" {
#  role   = aws_iam_role.processing_lambda_role.name
#  policy = data.aws_iam_policy_document.processing_lambda_policy.json
#}
#
#resource "aws_iam_role_policy_attachment" "processing_lambda_sqs_role_policy" {
#  role       = aws_iam_role.processing_lambda_role.name
#  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole"
#}