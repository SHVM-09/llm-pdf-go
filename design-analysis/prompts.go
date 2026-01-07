package main

// GeneratePrompt creates a comprehensive prompt based on output level
func GeneratePrompt(outputLevel string) string {
	basePrompt := `You are an expert CAD/Design document analyst. Analyze this PDF document which contains technical drawings, specifications, and design information.

Extract and organize the following information:

1. **DOCUMENT METADATA & APPROVAL INFORMATION**
   - Drawn By (name of designer/drafter)
   - Checked By (name of checker/reviewer)
   - Approved By (name of approver)
   - Dates (drawing date, check date, approval date)
   - Computer Code / CAD System information
   - Drawing projection type (e.g., Third Angle Projection)
   - Any revision history or version information

2. **PROJECT OVERVIEW**
   - Product/Component name
   - Drawing number and revision (for each drawing/page)
   - General description
   - Key dimensions (max length, width, height)
   - Weight if specified
   - Material designation (if specified at top level)

3. **BILL OF MATERIALS (BOM)**
   - Complete list of all parts with part numbers
   - Quantities for each part (exact counts)
   - Material specifications for each component
   - Component descriptions
   - Item numbers in sequence

4. **TECHNICAL SPECIFICATIONS**
   - All dimensions and tolerances (extract every dimension visible)
   - Material requirements (for each component)
   - Finish requirements (for each component)
   - Assembly notes and instructions
   - Manufacturing notes
   - General notes (e.g., "REMOVE ALL SHARP CORNERS", "DO NOT SCALE THIS DRAWING")
   - Unit of measurement (MM, inches, etc.)

5. **DRAWING INFORMATION**
   - Views available (front, side, top, 3D, exploded, section views)
   - Drawing standards used
   - Scale information
   - Critical dimensions
   - All geometric features (radii, angles, chamfers)
   - Drawing numbers for each component drawing

6. **ASSEMBLY INFORMATION**
   - Assembly sequence if available
   - Key assembly points
   - Component relationships
   - Exploded view details if present

7. **QUALITY & MANUFACTURING NOTES**
   - Special instructions
   - Quality requirements
   - Testing requirements
   - Any warnings or cautions

IMPORTANT: Extract ALL information from EVERY page, including title blocks, revision blocks, and notes sections. Do not skip any metadata fields like "Drawn By", "Checked By", "Approved By", dates, or computer codes.

Format your response clearly with sections and use markdown formatting for better readability.`

	switch outputLevel {
	case "executive":
		return basePrompt + `

**OUTPUT FORMAT (Executive Level):**
- Start with document metadata section (Drawn By, Checked By, Approved By, dates)
- Provide a concise executive summary (2-3 paragraphs)
- Highlight key components and their quantities from BOM
- Mention critical dimensions and specifications (overall size)
- Include any important manufacturing or quality notes
- Keep technical jargon minimal, focus on business impact
- Use bullet points for clarity
- Include drawing numbers and revision information`

	case "technical":
		return basePrompt + `

**OUTPUT FORMAT (Technical Level):**
- Start with complete document metadata section (Drawn By, Checked By, Approved By, dates, computer codes, projection type)
- Provide detailed technical information for ALL components
- Include complete BOM with all part numbers, quantities, and materials
- List ALL dimensions and tolerances from every drawing
- Extract geometric features (radii, angles, chamfers) for each component
- Explain technical specifications in detail for each part
- Include assembly and manufacturing details
- Document all general notes and special instructions
- Include drawing numbers and revisions for each component
- Use technical terminology appropriately
- Organize information in clear sections with headers
- Ensure NO technical details are omitted`

	case "detailed":
		return basePrompt + `

**OUTPUT FORMAT (Detailed Level):**
- Start with comprehensive document metadata section (Drawn By, Checked By, Approved By, all dates, computer codes, projection type, revision history)
- Provide comprehensive analysis of ALL aspects from EVERY page
- Include complete BOM with full descriptions, part numbers, quantities, materials
- Extract ALL dimensions, tolerances, and specifications from every drawing
- Document ALL geometric features (radii, angles, chamfers, fillets) with their values
- Detail all views and drawing information (front, side, top, 3D, exploded)
- Explain assembly procedures if visible
- Include ALL manufacturing and quality notes
- Extract ALL tables, notes, annotations, and title block information
- Provide page-by-page breakdown with drawing numbers and revisions
- Include all general notes (e.g., "REMOVE ALL SHARP CORNERS", scaling instructions)
- Document material and finish specifications for each component
- Use structured format with clear hierarchy
- Ensure COMPLETE extraction - nothing should be missed`

	default:
		return basePrompt + `

**OUTPUT FORMAT:**
- Provide a well-structured analysis
- Include all key information
- Use clear sections and formatting
- Make it readable for technical audiences`
	}
}
