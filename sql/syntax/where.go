package syntax

import (
	"errors"
)

//Unensurable
func whereParser(tr *TokenReader) (*SyntaxTreeNode, error) {
	t := tr.Read()
	if t.Kind != "keyword" && string(t.Raw) != "where" {
		return nil, errors.New("You have a syntax error near :" + string(t.Raw))
	}
	logical, err := orParser(tr)
	if err != nil {
		return nil, err
	}
	return &SyntaxTreeNode{
		Name:  "where",
		Child: []*SyntaxTreeNode{logical},
	}, nil
}

//duplicated
func logicalParser(tr *TokenReader) (*SyntaxTreeNode, error) {
	fork := tr.Fork()
	t := fork.Read()
	if t.Kind == "logical" {
		tr.Next(1)
		if string(t.Raw) != "not" {
			return nil, errors.New("You have a syntax error near :" + string(t.Raw))
		}
		logical, err := logicalParser(tr)
		if err != nil {
			return nil, err
		}
		return &SyntaxTreeNode{
			Name:  "logical",
			Value: t.Raw,
			Child: []*SyntaxTreeNode{logical},
		}, nil
	}
	fork = tr.Fork()
	relation, err := relationParser(fork)
	if err == nil {
		fork2 := fork.Fork()
		t := fork2.Read()
		if t.Kind != "logical" || (string(t.Raw) != "and" && string(t.Raw) != "or") {
			tr.pos = fork.pos
			return relation, nil
		}
		after, err := logicalParser(fork2)
		if err == nil {
			tr.pos = fork2.pos
			return &SyntaxTreeNode{
				Name:  "logical",
				Value: t.Raw,
				Child: []*SyntaxTreeNode{relation, after},
			}, nil
		} else {
			return nil, err
		}
	}
	if t.Kind == "structs" && string(t.Raw) == "(" {
		tr.Next(1)
		fork := tr.Fork()
		logical, err := logicalParser(fork)
		if err != nil {
			return nil, err
		}
		t := fork.Read()
		if t.Kind != "structs" || string(t.Raw) != ")" {
			return nil, errors.New("You have a syntax error near:" + string(t.Raw))
		}
		relation := &SyntaxTreeNode{
			Name: "()",
			Child: []*SyntaxTreeNode{
				logical,
			},
		}
		fork2 := fork.Fork()
		t = fork2.Read()
		if t.Kind != "logical" || (string(t.Raw) != "and" && string(t.Raw) != "or") {
			tr.pos = fork.pos
			return relation, nil
		}
		after, err := logicalParser(fork2)
		if err == nil {
			tr.pos = fork2.pos
			return &SyntaxTreeNode{
				Name:  "logical",
				Value: t.Raw,
				Child: []*SyntaxTreeNode{relation, after},
			}, nil
		} else {
			return nil, err
		}

	}
	return nil, errors.New("You have a syntax error near:" + string(t.Raw))
}

//Ensuable
func notParser(tr *TokenReader) (*SyntaxTreeNode, error) {
	fork := tr.Fork()
	t := fork.Read()
	if t.Kind == "logical" && string(t.Raw) == "not" {
		relation, err := relationParser(fork)
		if err != nil {
			return nil, err
		}
		tr.pos = fork.pos
		return &SyntaxTreeNode{
			Name:  "logical",
			Value: t.Raw,
			Child: []*SyntaxTreeNode{relation},
		}, nil
	}
	fork = tr.Fork()
	relation, err := relationParser(fork)
	if err != nil {
		return nil, err
	}
	tr.pos = fork.pos
	return relation, nil
}

//Ensuable
func relationParser(tr *TokenReader) (*SyntaxTreeNode, error) {
	fork := tr.Fork()
	t := fork.Read()
	if t.Kind == "structs" && string(t.Raw) == "(" {
		or, err := orParser(fork)
		if err == nil {
			t = fork.Read()
			if t.Kind == "structs" && string(t.Raw) == ")" {
				tr.pos = fork.pos
				return or, nil
			}
		}
	}
	fork = tr.Fork()
	exp1, err := expressionParser(fork)
	if err != nil {
		return nil, err
	}
	t = fork.Read()
	if t.Kind != "relations" {
		return nil, errors.New("You have a syntax error near:" + string(t.Raw))
	}
	exp2, err := expressionParser(fork)
	if err != nil {
		return nil, err
	}
	tr.pos = fork.pos
	return &SyntaxTreeNode{
		Name:  "relations",
		Value: t.Raw,
		Child: []*SyntaxTreeNode{exp1, exp2},
	}, nil
}

//Ensurable
func orParser(tr *TokenReader) (*SyntaxTreeNode, error) {
	fork := tr.Fork()
	and, err := andParser(fork)
	if err != nil {
		return nil, err
	}
	fork2 := fork.Fork()
	t := fork2.Read()
	if t.Kind != "logical" || string(t.Raw) != "or" {
		tr.pos = fork.pos
		return and, nil
	}
	or, err := orParser(fork2)
	if err != nil {
		return nil, err
	}
	tr.pos = fork2.pos
	return &SyntaxTreeNode{
		Name:  "logical",
		Value: t.Raw,
		Child: []*SyntaxTreeNode{and, or},
	}, nil
}

//Ensurable
func andParser(tr *TokenReader) (*SyntaxTreeNode, error) {
	fork := tr.Fork()
	relation, err := notParser(fork)
	if err != nil {
		return nil, err
	}
	fork2 := fork.Fork()
	t := fork2.Read()
	if t.Kind != "logical" || string(t.Raw) != "and" {
		tr.pos = fork.pos
		return relation, nil
	}
	and, err := andParser(fork2)
	if err != nil {
		return nil, err
	}
	tr.pos = fork2.pos
	return &SyntaxTreeNode{
		Name:  "logical",
		Value: t.Raw,
		Child: []*SyntaxTreeNode{relation, and},
	}, nil
}

//Ensurable
func expressionParser(tr *TokenReader) (*SyntaxTreeNode, error) {
	fork := tr.Fork()
	t := fork.Read()
	if t.Kind == "identical" {
		tr.Next(1)
		return &SyntaxTreeNode{
			Name:      "identical",
			Value:     t.Raw,
			ValueType: NAME,
		}, nil
	}
	fork = tr.Fork()
	n, err := referParser(fork)
	if err == nil {
		tr.pos = fork.pos
		return n, nil
	}
	fork = tr.Fork()
	n, err = filedParser(fork)
	if err == nil {
		tr.pos = fork.pos
		return n, nil
	}
	fork = tr.Fork()
	n, err = valueParser(fork)
	if err == nil {
		tr.pos = fork.pos
		return n, nil
	}
	return nil, errors.New("You have a syntax error near : " + string(t.Raw))
}
