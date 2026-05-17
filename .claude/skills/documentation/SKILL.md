---
name: go-doc
description: Write or improve Go documentation — package comments, godoc, README, and API docs. Trigger on any request to document, add comments, write godoc, or improve existing Go docs.
---


Write idiomatic Go documentation: package comments start with "Package X ...", exported symbols get godoc comments starting with the symbol name, examples use `Example_()` test functions. For README, include: what it does, install, quickstart, and key API surface. For internal packages, document non-obvious behavior, concurrency guarantees, and error conditions. Output ready-to-paste doc blocks — no filler, no "This function does X" restatements of the signature.

Every exported function and method MUST have a doc comment. Document complex internal functions too. Skip test functions.


The comment starts with the function name and a verb phrase. Focus on why and when, not restating what the code already shows. The code tells you what happens — the comment should explain why it exists, when to use it, what constraints apply, and what can go wrong. Include parameters, return values, error cases, and a usage example:

// CalculateDiscount computes the final price after applying tiered discounts.
// Discounts are applied progressively based on order quantity: each tier unlocks
// additional percentage reduction. Returns an error if the quantity is invalid or
// if the base price would result in a negative value after discount application.
//
// Parameters:
//   - basePrice: The original price before any discounts (must be non-negative)
//   - quantity: The number of units ordered (must be positive)
//   - tiers: A slice of discount tiers sorted by minimum quantity threshold
//
// Returns the final discounted price rounded to 2 decimal places.
// Returns ErrInvalidPrice if basePrice is negative.
// Returns ErrInvalidQuantity if quantity is zero or negative.
//
// Play: https://go.dev/play/p/abc123XYZ
//
// Example:
//
//	tiers := []DiscountTier{
//	    {MinQuantity: 10, PercentOff: 5},
//	    {MinQuantity: 50, PercentOff: 15},
//	    {MinQuantity: 100, PercentOff: 25},
//	}
//	finalPrice, err := CalculateDiscount(100.00, 75, tiers)
//	if err != nil {
//	    log.Fatalf("Discount calculation failed: %v", err)
//	}
//	log.Printf("Ordered 75 units at $100 each: final price = $%.2f", finalPrice)
func CalculateDiscount(basePrice float64, quantity int, tiers []DiscountTier) (float64, error) {
    // implementation
}