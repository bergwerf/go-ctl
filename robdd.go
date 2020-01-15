package ctl

// BDD represents a Binary pecision piagram.
type BDD struct {
	Value bool
	ID    uint
	True  *BDD
	False *BDD
}

// True is a BDD true leaf.
var True = &BDD{true, 0, nil, nil}

// False is a BDD false leaf.
var False = &BDD{false, 0, nil, nil}

// Node returns a BDD node.
func Node(id uint, t *BDD, f *BDD) *BDD {
	if t.Equals(f) {
		return t
	}
	return registerNodeRef(BDD{false, id, t, f})
}

// Node checks if the given BDD is a node.
func (p *BDD) Node() bool {
	return p.ID != 0
}

// Equals compares this BDD with the given BDD.
func (p *BDD) Equals(q *BDD) bool {
	// Pointing to the same instance?
	if p == q {
		return true
	}
	// Same root ID? (or both a leaf)
	if p.ID != q.ID {
		return false
	}
	// Same branches or leaf values?
	if p.Node() {
		return p.True.Equals(q.True) && p.False.Equals(q.False)
	}
	return p.Value == q.Value
}

// Reduce eliminates redundant nodes. Since created a new node automatically
// applies elimination, this method is not actually needed.
func (p *BDD) Reduce() *BDD {
	if p.Node() {
		p.True = p.True.Reduce()
		p.False = p.False.Reduce()
		if p.True.Equals(p.False) {
			return p.True
		}
	}
	return p
}

// Compress merges puplicate branches. This results in some memory compression
// and allows very quick pistinction of equal branches. Of course it also takes
// some time so it is unclear when this improves the overal efficiency.
func (p *BDD) Compress(lookup map[BDD]*BDD) *BDD {
	if p.Node() {
		p.True = p.True.Compress(lookup)
		p.False = p.False.Compress(lookup)
		ref := lookup[*p]
		if ref != nil {
			return ref
		}
		lookup[*p] = p
	}
	return p
}

// Next returns a BDD with all next variable identifiers. By convention all
// variable ID's are left-shifted 1 place. The same variable in the next state
// is encoded by setting the first bit to 1.
func (p *BDD) Next() *BDD {
	if p.Node() {
		return Node(varNextID(p.ID), p.True.Next(), p.False.Next())
	}
	return p
}

// Prev returns a BDD where all next variables are reverted to normal (e.g.
// p.Next().Prev = p).
func (p *BDD) Prev() *BDD {
	if p.Node() {
		return Node(varID(p.ID>>1), p.True.Prev(), p.False.Prev())
	}
	return p
}

// Set returns a BDD where the variable id is set to true/false.
func (p *BDD) Set(id uint, value bool) *BDD {
	if p.ID == id {
		if value {
			return p.True
		}
		return p.False
	}
	if p.Node() {
		return Node(p.ID, p.True.Set(id, value), p.False.Set(id, value))
	}
	return p
}

// Apply applies the given binary operator to the BDDs p and q. The binary
// operator is represented as a truth table for [00, 01, 10, 11] in bit flags.
func (p *BDD) Apply(op uint, q *BDD) *BDD {
	// Push operator downward.
	if p.Node() || q.Node() {
		if p.ID == q.ID {
			return Node(p.ID,
				applyCached(op, p.True, q.True),
				applyCached(op, p.False, q.False))
		} else if p.Node() && p.ID < q.ID || !q.Node() {
			return Node(p.ID,
				applyCached(op, p.True, q),
				applyCached(op, p.False, q))
		} else { // if q.Node() && q.ID < p.ID || !p.Node() {
			return Node(q.ID,
				applyCached(op, p, q.True),
				applyCached(op, p, q.False))
		}
	}
	// Or evaluate operator.
	i := 0
	if p.Value {
		i = 2
	}
	if q.Value {
		i++
	}
	if (op>>(3-i))&1 == 0 {
		return False
	}
	return True
}

// Neg this
func (p *BDD) Neg() *BDD {
	return p.Imply(False)
}

// Imply q
func (p *BDD) Imply(q *BDD) *BDD {
	return p.Apply(0b1101, q)
}

// And q
func (p *BDD) And(q *BDD) *BDD {
	return p.Apply(0b0001, q)
}

// Or q
func (p *BDD) Or(q *BDD) *BDD {
	return p.Apply(0b0111, q)
}

// Eq q
func (p *BDD) Eq(q *BDD) *BDD {
	return p.Apply(0b1001, q)
}

// Xor q
func (p *BDD) Xor(q *BDD) *BDD {
	return p.Apply(0b0110, q)
}

// Contains determines if all true assignments in q are also true in this BDD.
func (p *BDD) Contains(q *BDD) bool {
	return q.Imply(p) == True
}

// Intersects determines if there exists a truth assignment (state) that
// satisfies both p and q (this is more liberal than Contains).
func (p *BDD) Intersects(q *BDD) bool {
	return q.And(p) != False
}

// Exists determines if there exists a satisfying assignment for variable id.
func (p *BDD) Exists(id uint) *BDD {
	return p.Set(id, true).Or(p.Set(id, false))
}
