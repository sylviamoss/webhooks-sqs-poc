data "aws_iam_policy_document" "processing_lambda_policy" {
  // Allow processing lambda to invoke egress lambda.
  statement {
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction",
    ]
    resources = [aws_lambda_function.webhooks_egress_lambda.arn]
  }
  statement {
    effect = "Allow"
    actions = [
      "sqs:GetQueueUrl",
      "sqs:ChangeMessageVisibility",
    ]
    resources = [aws_sqs_queue.webhooks_sqs.arn]
  }
}

resource "aws_iam_role" "processing_lambda_role" {
  name = "processing-lambda-role"
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

resource "aws_iam_role_policy" "processing_lambda_role_policy" {
  role   = aws_iam_role.processing_lambda_role.name
  policy = data.aws_iam_policy_document.processing_lambda_policy.json
}

resource "aws_iam_role_policy_attachment" "processing_lambda_sqs_role_policy" {
  role       = aws_iam_role.processing_lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole"
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