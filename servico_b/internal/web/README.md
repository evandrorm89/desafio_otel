Desafio Open Telemetry para a pós Goexpert

Para rodar o projeto, basta ir para a raiz e rodar o seguinte comando: **docker-compose up --build** para subir os containers do serviço A e B e também o do zipkin com Open Telemetry

Os arquivos Docker buildam cada serviço juntamente com os testes unitários e de integração rodados

Para acessar o monitoramento do zipkin, acessar http://localhost:9411

Para realizar uma chamada, realizar um POST para http://localhost:8080, com o seguinte body

{
  "cep": "12345678"
}

O formato esperado da resposta caso o status seja 200 será o seguinte:

{ "city: "cidade", "temp_C": numero, "temp_F": numero, "temp_K": numero }
