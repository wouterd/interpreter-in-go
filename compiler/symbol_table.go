package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	FreeSymbols    []Symbol
	outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.outer = outer
	return s
}

func (t *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: t.numDefinitions,
	}
	if t.outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	t.store[name] = symbol
	t.numDefinitions++
	return symbol
}

func (t *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	t.store[name] = symbol
	return symbol
}

func (t *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 69420, Scope: FunctionScope}
	t.store[name] = symbol
	return symbol
}

func (t *SymbolTable) defineFree(orig Symbol) Symbol {
	t.FreeSymbols = append(t.FreeSymbols, orig)

	symbol := Symbol{Name: orig.Name, Index: len(t.FreeSymbols) - 1}
	symbol.Scope = FreeScope

	t.store[orig.Name] = symbol
	return symbol
}

func (t *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := t.store[name]
	if ok || t.outer == nil {
		return s, ok
	}

	s, ok = t.outer.Resolve(name)
	if !ok || s.Scope == GlobalScope || s.Scope == BuiltinScope {
		return s, ok
	}

	free := t.defineFree(s)
	return free, true
}
