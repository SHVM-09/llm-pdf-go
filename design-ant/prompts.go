package main

import "fmt"

// generateAnalysisPrompt creates the prompt for design analysis
func generateAnalysisPrompt(pageNumber int) string {
	return fmt.Sprintf(`Analyze this single PDF page completely. Extract ALL technical details, dimensions, parts, and specifications. DO NOT skip, omit, or summarize anything.

OUTPUT FORMAT - START DIRECTLY (NO INTRODUCTORY PHRASES):
Start your response immediately with:
# Page %d

Then provide the analysis in the following structure:

1. **METADATA**: Drawn By, Checked By, Approved By (exact names), dates, drawing numbers, revisions, CAD codes, projection type

2. **OVERVIEW**: Component name, description, key dimensions (with units), weight, material codes

3. **BOM**: List EVERY part number (P01, P02, etc.) - extract ALL rows from tables. Include quantities, materials, descriptions. State total part count.

4. **DIMENSIONS**: ALL linear, diameter (Ø), radius (R), angles, distances, depths. Include tolerances. Format: [Feature]: [Value] [Unit]

5. **DRAWINGS**: All views (front/side/top/3D/exploded/section), scales, standards. ALL geometric features: radii, angles, chamfers, fillets, threads with exact values

6. **ASSEMBLY**: Sequence, assembly points, relationships, fastening methods, tolerances

7. **NOTES**: Manufacturing, quality, testing, warnings, inspection requirements - EXACT text

8. **MATERIALS/FINISHES**: Exact codes for each component

CRITICAL RULES:
- DO NOT write "Here's a comprehensive extraction..." or "I'll extract..." or any introductory phrases
- DO NOT write "Let me analyze..." or similar phrases
- Start immediately with: # Page %d
- List EVERY part, dimension, and component - no "etc." or "various"
- Extract EXACT values - no approximations
- If table has 25 rows, list all 25
- If exploded view shows 20 parts, list all 20
- Use tables/numbered lists for clarity

BEGIN NOW - Start with page number and heading:`, pageNumber, pageNumber)
}
