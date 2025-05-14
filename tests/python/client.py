import requests
import json
import errors
import time

class Calculator:


    def __init__(self, endpoint: str):
        self.endpoint = endpoint

    def post_request(self, path: str, body: dict, token=None):
        # Выполняет HTTP запрос к API оркестратора с заданными параметрами
        response = None

        if token is not None:
            # authorizing
            response = requests.post(
                url=self.endpoint+path,
                data=json.dumps(body),
                headers={
                    "Authorization": f"Bearer {token}"
                }
            )
        else:
            response = requests.post(
                url=self.endpoint+path,
                data=json.dumps(body)
            )

        if response.status_code == 400:
            raise errors.BadRequestException(response_body=response.text)
        elif response.status_code == 401:
            raise errors.UnauthorizedException(response_body=response.text)
        elif response.status_code == 404:
            raise errors.NotFoundException(response_body=response.text)
        elif response.status_code == 500:
            raise errors.InternalServerErrorException(response_body=response.text) 
        
        return response 

    def get_request(self, path: str, token=None):
        # Выполняет HTTP запрос к API оркестратора с заданными параметрами
        response = None

        if token is not None:
            # authorizing
            response = requests.get(
                url=self.endpoint+path,
                headers={
                    "Authorization": f"Bearer {token}"
                }
            )
        else:
            response = requests.get(
                url=self.endpoint+path,
            )

        if response.status_code == 400:
            raise errors.BadRequestException(response_body=response.text)
        elif response.status_code == 401:
            raise errors.UnauthorizedException(response_body=response.text)
        elif response.status_code == 404:
            raise errors.NotFoundException(response_body=response.text)
        elif response.status_code == 500:
            raise errors.InternalServerErrorException(response_body=response.text) 
        
        return response 
               
    def register(self, login: str, password: str) -> None:
        # Регистрирует нового пользователя и возвращает токен
        response = self.post_request(path="/register", body={
            "login": login,
            "password": password
        })
    
    def login(self, login: str, password: str) -> str:
        # Выполняет вход пользователя и возвращает токен
        response = self.post_request(path="/login", body={
            "login": login,
            "password": password
        })  
             
        json_response = response.json()
        return json_response["token"]

    def calculate(self, expression: str, token: str) -> str:
        # Отправляет выражение на вычисление и ожидает результат
        response = self.post_request(path="/calculate", body={
            "expression": expression
        }, token=token)

        json_response = response.json()
        expr_id = json_response["id"]

        return expr_id

    def expression_by_id(self, id: str, token:str) -> float | None:
        # Получает результат вычисления выражения по его идентификатору
        while True:
            try:
                response = self.get_request(path="/expressions/"+id, token=token)
                break
            except:
                time.sleep(0.1)
                
        json_response = response.json()

        if json_response["expression"]["status"] == "COMPLETE":
            return json_response["expression"]["result"]
        else:
            return None
    
    # def expressions(self, token: str) -> List[float] | None: