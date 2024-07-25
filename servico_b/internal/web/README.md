Desafio Open Telemetry para a pós Goexpert

Para rodar o projeto, basta ir para a raiz e rodar docker-compose up --build para subir os containers do serviço A e B e também o do zipkin com Open Telemetry

Os arquivos Docker buildam cada serviço juntamente com os testes unitários e de integração rodados

Para acessar o monitoramento do zipkin, acessar http://localhost:9411

Para realizar uma chamada, realizar um POST para http://localhost:8080, com o seguinte body

{
  "cep": "<string>"
}

O formato esperado da resposta caso o status seja 200 será o seguinte:

{ "city: "<string>", "temp_C": <number>, "temp_F": <number>, "temp_K": <number> }
