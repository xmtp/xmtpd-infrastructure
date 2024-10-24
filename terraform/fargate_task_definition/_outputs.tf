output "task_definition_arn" {
  description = "The ARN of the task defintion"
  value       = aws_ecs_task_definition.task.arn
}