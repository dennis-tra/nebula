output "crawler_ip" {
    description = "Public IP address of crawler"
    value = aws_instance.ipfs-crawler.public_ip
}