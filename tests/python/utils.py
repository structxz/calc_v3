import random
import string

class Counter:
    all_count = 0
    passed_count = 0

    def add(self):
        self.all_count += 1
    
    def passed(self):
        self.passed_count += 1

    def final(self):
        return (self.passed_count, self.all_count)

def generate_random_string(length):
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for _ in range(length))

def bold(text):
    print(f"\033[1m{text}\033[0m")

def pass_(text):
    print(f"\033[32m{text}\033[0m")

def fail(text):
    print(f"\033[31m{text}\033[0m")

def part(text):
    print(f"\033[33m{text}\033[0m")