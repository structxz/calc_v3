from client import Calculator
from utils import pass_, fail, bold, Counter, generate_random_string


def registration_test() -> None:
    c: Counter = Counter()
    # Test 1: Correct registration
    try:
        c.add()
        calc.register(login=test_login, password="password123")
        pass_("Test 1 passed: Successful registration")
        c.passed()
    except Exception as e:
        fail(f"Test 1 did not passed: {e}")
        return

    # Test 2: Empty name of user
    try:
        c.add()
        calc.register(login="", password="password123")
        fail("Test 2 did not passed: Не должно допускаться пустое имя пользователя")
    except Exception as e:
        pass_("Test 2 passed: Empty login rejected")
        c.passed()

    # Test 3: Repeated registration
    try:
        c.add()
        calc.register(login=test_login, password="password123")
        fail("Test 3 did not passed: Repeated registration is not allowed")
    except Exception as e:
        pass_("Test 3 passed: Repeated registration rejected")
        c.passed()

    print(f"Passed: ({c.final()[0]}/{c.final()[1]})\n")

def login_test() -> None:
    c: Counter = Counter()
    # 1: Correct authorization
    try:
        c.add()
        token = calc.login(login=test_login, password="password123")
        pass_("Test 1 passed: Successful authorization")
        c.passed()
    except Exception as e:
        fail(f"Test 1 did not passed: {e}")
        return

    # Test 2: wrong password
    try:
        c.add()
        calc.login(login=test_login, password="wrongpassword")
        fail("Test 2 did not passed: Authorization with the wrong password is not allowed")
    except Exception as e:
        pass_("Test 2 passed: Authorization with incorrect password rejected")
        c.passed()

    # Test 3: Non-existent user
    try:
        c.add()
        calc.login(login="nonexistentuser", password="password123")
        fail("Test 3 did not passed: Authorization of a non-existent user is allowed")
    except Exception as e:
        pass_("Test 3 passed: Authorization of a non-existent user rejected")
        c.passed()

    print(f"Passed: ({c.final()[0]}/{c.final()[1]})\n")

def calculate_test() -> None:
    c: Counter = Counter()
    try:
        c.add()
        token: str = calc.login(login=test_login, password="password123")
        id: str = calc.calculate(expression="2+2", token=token)
        if id != "" and id is not None:
            c.passed()
            pass_("Test 1 passed: Successful sending for calculation")
        else:
            fail("Test 1 did not passed: empty ID")
    except Exception as e:
        fail(f"Test 1 did not passed: {e}")
        return
    print(f"Passed: ({c.final()[0]}/{c.final()[1]})\n")

def expression_test() -> None:
    c: Counter = Counter()
    # Test 1: correct
    try:
        c.add()
        token: str = calc.login(login=test_login, password="password123")
        id: str = calc.calculate(expression="2+2", token=token)
        result = calc.expression_by_id(id=id, token=token)
        if result == float(4):
            c.passed()
            pass_("Test 1 passed: Calculation successful")
        else:
            fail(f"Test 1 did not passed: Wrong answer: {result}")
    except Exception as error:
        fail(f"Test 1 did not passed: {error}")
        return
    
    print(f"Passed: ({c.final()[0]}/{c.final()[1]})\n")


if __name__ == "__main__":
    ENDPOINT = "http://localhost:8080/api/v1"
    calc = Calculator(ENDPOINT)

    test_login = generate_random_string(8)
    
    bold("Registration:")
    registration_test()

    bold("Authorization:")
    login_test()

    bold("Sending for calculation:")
    calculate_test()

    bold("Get expression result:")
    expression_test()
