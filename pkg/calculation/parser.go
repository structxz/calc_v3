package calculation

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"distributed_calculator/internal/constants"
	"go.uber.org/zap"
)

// Parser represents a mathematical expression parser.
type Parser struct {
	tokens []string // Tokens of the expression to be parsed.
	pos    int      // Current position in the tokens slice.
}

// parse evaluates the entire expression and returns the result.
// It ensures that all tokens are consumed and returns an error if unexpected tokens remain.
func (p *Parser) parse() (float64, error) {
	result, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	if p.pos < len(p.tokens) {
		return 0, errors.New(constants.ErrUnexpectedToken)
	}
	return result, nil
}

// parseExpression parses addition and subtraction operations.
func (p *Parser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.tokens) {
		op := p.tokens[p.pos]
		if op != "+" && op != "-" {
			break
		}
		p.pos++

		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}

		if op == "+" {
			left += right
		} else {
			left -= right
		}
	}

	return left, nil
}

// parseTerm parses multiplication, division, and modulo operations.
func (p *Parser) parseTerm() (float64, error) {
	left, err := p.parsePower()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.tokens) {
		op := p.tokens[p.pos]
		if op != "*" && op != "/" && op != "%" {
			break
		}
		p.pos++

		right, err := p.parsePower()
		if err != nil {
			return 0, err
		}

		switch op {
		case "*":
			left *= right
		case "/":
			if right == 0 {
				return 0, errors.New(constants.ErrDivisionByZero)
			}
			left /= right
		case "%":
			if right == 0 {
				return 0, errors.New(constants.ErrModuloByZero)
			}
			if left != float64(int(left)) || right != float64(int(right)) {
				return 0, errors.New(constants.ErrInvalidModulo)
			}
			left = math.Mod(left, right)
		}
	}

	return left, nil
}

// parsePower parses exponentiation operations.
func (p *Parser) parsePower() (float64, error) {
	result, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	if p.pos < len(p.tokens) && p.tokens[p.pos] == "^" {
		p.pos++

		exponent, err := p.parsePower()
		if err != nil {
			return 0, err
		}
		result = math.Pow(result, exponent)
	}

	return result, nil
}

// parseFactor parses individual factors, including numbers, parentheses, and negative signs.
func (p *Parser) parseFactor() (float64, error) {
	if p.pos >= len(p.tokens) {
		if logger != nil {
			logger.Error(constants.LogUnexpectedEndExpr,
				zap.Strings(constants.FieldTokens, p.tokens),
				zap.Int(constants.FieldPosition, p.pos))
		}
		return 0, errors.New(constants.ErrUnexpectedEndExpr)
	}

	token := p.tokens[p.pos]
	p.pos++

	switch {
	case token == "(":
		result, err := p.parseExpression()
		if err != nil {
			if logger != nil {
				logger.Error(constants.LogFailedParseParentheses,
					zap.Error(err),
					zap.Strings(constants.FieldTokens, p.tokens),
					zap.Int(constants.FieldPosition, p.pos))
			}
			return 0, err
		}
		if p.pos >= len(p.tokens) || p.tokens[p.pos] != ")" {
			if logger != nil {
				logger.Error(constants.LogMissingCloseParen,
					zap.Strings(constants.FieldTokens, p.tokens),
					zap.Int(constants.FieldPosition, p.pos))
			}
			return 0, errors.New(constants.ErrMissingCloseParen)
		}
		p.pos++
		return result, nil
	case token == "-":
		factor, err := p.parseFactor()
		if err != nil {
			if logger != nil {
				logger.Error(constants.LogFailedParseNegative,
					zap.Error(err),
					zap.Strings(constants.FieldTokens, p.tokens),
					zap.Int(constants.FieldPosition, p.pos))
			}
			return 0, err
		}
		return -factor, nil
	case isNumber(token):
		num, err := strconv.ParseFloat(token, 64)
		if err != nil {
			if logger != nil {
				logger.Error(constants.LogInvalidNumberFormat,
					zap.String(constants.FieldToken, token),
					zap.Error(err))
			}
			return 0, fmt.Errorf("invalid number: %s", token)
		}
		return num, nil
	default:
		if logger != nil {
			logger.Error(constants.LogUnexpectedToken,
				zap.String(constants.FieldToken, token),
				zap.Strings(constants.FieldTokens, p.tokens),
				zap.Int(constants.FieldPosition, p.pos))
		}
		return 0, fmt.Errorf("unexpected token: %s", token)
	}
}
