class BadRequestException(Exception):
    def __init__(self, message="Bad request", response_body=None):
        self.message = message
        self.response_body = response_body
        super().__init__(f"{self.message}, Response body: {self.response_body}")

class UnauthorizedException(Exception):
    def __init__(self, message="Unauthorized", response_body=None):
        self.message = message
        self.response_body = response_body
        super().__init__(f"{self.message}, Response body: {self.response_body}")

class InternalServerErrorException(Exception):
    def __init__(self, message="Internal server error", response_body=None):
        self.message = message
        self.response_body = response_body
        super().__init__(f"{self.message}, Response body: {self.response_body}")

class NotFoundException(Exception):
    def __init__(self, message="Not found", response_body=None):
        self.message = message
        self.response_body = response_body
        super().__init__(f"{self.message}, Response body: {self.response_body}")