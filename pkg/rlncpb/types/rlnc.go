package types

// NewSymbolSet creates a reference to a new SymbolSet with the provided symbols.
func NewSymbolSet(symbols ...*Symbol) *SymbolSet {
	if len(symbols) == 0 {
		return &SymbolSet{
			Symbols: make([]*Symbol, 0),
		}
	}

	return &SymbolSet{
		Symbols: symbols,
	}
}

// Size returns the number of symbols in the SymbolSet. If the SymbolSet is nil,
// it returns 0.
func (ss *SymbolSet) Size() int {
	if ss == nil {
		return 0
	}

	return len(ss.GetSymbols())
}

// K returns the coefficient-vector width for the SymbolSet.
//
// In a valid coding window, every symbol has the same coefficient width, so K is
// read from the first symbol. K returns 0 for nil, empty, or structurally invalid
// SymbolSets.
func (ss *SymbolSet) K() int {
	if ss == nil || len(ss.GetSymbols()) == 0 || ss.GetSymbols()[0] == nil {
		return 0
	}

	return len(ss.GetSymbols()[0].GetCoefficients())
}
