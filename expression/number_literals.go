package expression

import (
	errors2 "github.com/terawatthour/socks/errors"
	"strconv"
	"strings"
)

func (p *parser) parseNumeric() (Expression, error) {
	if strings.Contains(p.currentToken.Literal, ".") {
		return p.parseDecimalFloat()
	}
	if strings.HasPrefix(p.currentToken.Literal, "0x") {
		return p.parseHex()
	}
	if strings.HasPrefix(p.currentToken.Literal, "0b") {
		return p.parseBinary()
	}
	if strings.HasPrefix(p.currentToken.Literal, "0") && len(p.currentToken.Literal) > 1 {
		return p.parseOctal()
	}
	return p.parseDecimalInteger()
}

func (p *parser) parseOctal() (Expression, error) {
	lit := strings.TrimPrefix(p.currentToken.Literal[1:], "c")
	i64, err := strconv.ParseInt(lit, 8, 64)
	if err != nil {
		return nil, errors2.New("could not parse `"+p.currentToken.Literal+"` as octal", p.currentToken.Location)
	}

	return &Integer{Token: p.currentToken, Value: int(i64)}, nil
}

func (p *parser) parseHex() (Expression, error) {
	i64, err := strconv.ParseInt(p.currentToken.Literal[2:], 16, 64)
	if err != nil {
		return nil, errors2.New("could not parse `"+p.currentToken.Literal+"` as hex", p.currentToken.Location)
	}

	return &Integer{Token: p.currentToken, Value: int(i64)}, nil
}

func (p *parser) parseBinary() (Expression, error) {
	i64, err := strconv.ParseInt(p.currentToken.Literal[2:], 2, 64)
	if err != nil {
		return nil, errors2.New("could not parse `"+p.currentToken.Literal+"` as binary", p.currentToken.Location)
	}

	return &Integer{Token: p.currentToken, Value: int(i64)}, nil
}

func (p *parser) parseDecimalInteger() (Expression, error) {
	res, err := strconv.ParseInt(p.currentToken.Literal, 10, 64)
	if err != nil {
		return nil, errors2.New("could not parse `"+p.currentToken.Literal+"` as integer", p.currentToken.Location)
	}

	return &Integer{Token: p.currentToken, Value: int(res)}, nil
}

func (p *parser) parseDecimalFloat() (Expression, error) {
	res, err := strconv.ParseFloat(p.currentToken.Literal, 64)
	if err != nil {
		return nil, errors2.New("could not parse `"+p.currentToken.Literal+"` as float", p.currentToken.Location)
	}

	return &Float{Token: p.currentToken, Value: res}, nil
}