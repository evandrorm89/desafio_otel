Consulta de CEP e clima atual da localidade com open telemetry

Instruções:

- Na pasta raiz do projeto, rodar o comando docker-compose up --build
- Assim que todos os serviços subirem, fazer uma requisição POST para http://localhost:8080 com o seguinte body:

{
    "cep": ""
}

Dentro da string vazia, preencher o cep em que se deseja consultar a temperatura atual.

A resposta da API vem no seguinte formato de exemplo:

{
    "city": "São Paulo",
    "temp_C": 28.5
    "temp_F": 78.5
    "temp_K": 273.5
}

Para verificar os logs do open telemetry, acessar localhost:9411, onde está rodando o serviço zipkin
