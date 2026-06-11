package parser

import (
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/token"
)

// parseExpr parses one expression. The expression sublanguage is total: the
// only branching is the if/else-if/else conditional, and the only binary
// operators are concatenation (+) and path-join (/). Comparisons (== != =~)
// appear only inside conditional conditions.
func (p *parser) parseExpr() ast.Expr {
	if p.curKind() == token.IF {
		return p.parseConditional()
	}
	return p.parseConcat()
}

// parseConcat handles left-associative + (concat) and / (path-join), which
// share precedence.
func (p *parser) parseConcat() ast.Expr {
	left := p.parsePrimary()
	for p.curKind() == token.PLUS || p.curKind() == token.SLASH {
		op := p.advance()
		right := p.parsePrimary()
		if left == nil || right == nil {
			return left
		}
		left = &ast.Binary{Op: op.Kind, Left: left, Right: right, Sp: left.Span().To(right.Span())}
	}
	return left
}

func (p *parser) parsePrimary() ast.Expr {
	switch p.curKind() {
	case token.STRING:
		t := p.advance()
		return &ast.StringLit{Value: t.Lit, Sp: t.Span}
	case token.IDENT:
		t := p.advance()
		if p.curKind() == token.LPAREN {
			return p.parseFuncCall(t)
		}
		return &ast.VarRef{Name: t.Lit, Sp: t.Span}
	case token.LPAREN:
		p.advance()
		e := p.parseExpr()
		p.expect(token.RPAREN, "to close grouped expression")
		return e
	case token.IF:
		return p.parseConditional()
	default:
		// Consume the unexpected token (error recovery): every loop that parses a
		// sequence of expressions relies on parsePrimary making progress. Returning
		// without advancing here let callers (e.g. dependency/func-call arguments)
		// spin forever, appending nodes until OOM. advance() is a no-op at EOF.
		t := p.advance()
		p.errorf(t.Span, "expected an expression, found %s", describe(t))
		return &ast.StringLit{Value: "", Sp: t.Span}
	}
}

func (p *parser) parseFuncCall(name token.Token) ast.Expr {
	lp := p.advance() // LPAREN
	call := &ast.FuncCall{Name: name.Lit, Sp: name.Span.To(lp.Span)}
	if p.curKind() != token.RPAREN {
		call.Args = append(call.Args, p.parseExpr())
		for p.curKind() == token.COMMA {
			p.advance()
			call.Args = append(call.Args, p.parseExpr())
		}
	}
	rp, _ := p.expect(token.RPAREN, "to close function call")
	call.Sp = name.Span.To(rp.Span)
	return call
}

func isCmpOp(k token.Kind) bool {
	return k == token.EQ || k == token.NEQ || k == token.MATCH
}

func (p *parser) parseConditional() ast.Expr {
	ifTok := p.advance() // IF
	cond := &ast.Conditional{Sp: ifTok.Span}

	parseBranch := func() (ast.CondBranch, bool) {
		left := p.parseConcat()
		if !isCmpOp(p.curKind()) {
			p.errorf(p.cur().Span, "expected a comparison operator (== != =~), found %s", describe(p.cur()))
			return ast.CondBranch{}, false
		}
		op := p.advance()
		right := p.parseConcat()
		if _, ok := p.expect(token.LBRACE, "to open conditional result"); !ok {
			return ast.CondBranch{}, false
		}
		result := p.parseExpr()
		p.expect(token.RBRACE, "to close conditional result")
		return ast.CondBranch{Left: left, Op: op.Kind, Right: right, Result: result}, true
	}

	if b, ok := parseBranch(); ok {
		cond.Branches = append(cond.Branches, b)
	}

	for p.curKind() == token.ELSE {
		p.advance() // ELSE
		if p.curKind() == token.IF {
			p.advance() // IF (else if)
			if b, ok := parseBranch(); ok {
				cond.Branches = append(cond.Branches, b)
			}
			continue
		}
		// Final else.
		if _, ok := p.expect(token.LBRACE, "to open else result"); ok {
			cond.Else = p.parseExpr()
			rb, _ := p.expect(token.RBRACE, "to close else result")
			cond.Sp = ifTok.Span.To(rb.Span)
		}
		return cond
	}

	p.errorf(p.cur().Span, "conditional requires a final 'else { ... }'")
	if cond.Else == nil {
		cond.Else = &ast.StringLit{Value: "", Sp: ifTok.Span}
	}
	return cond
}

// parseExprList parses a [ expr, expr, ... ] list (settings & cache specs).
func (p *parser) parseExprList() []ast.Expr {
	if _, ok := p.expect(token.LBRACK, "to open a list"); !ok {
		return nil
	}
	var list []ast.Expr
	if p.curKind() != token.RBRACK {
		list = append(list, p.parseExpr())
		for p.curKind() == token.COMMA {
			p.advance()
			if p.curKind() == token.RBRACK {
				break // trailing comma
			}
			list = append(list, p.parseExpr())
		}
	}
	p.expect(token.RBRACK, "to close a list")
	return list
}
