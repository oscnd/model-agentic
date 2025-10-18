# Claude

General guideline:
- Use pointer for struct.
- Use r as the receiver name. Example: `func (r *Handler) HandleOrganizationCreate(c *fiber.Ctx) error`.
- Comment in format of `// * lowercase compact action` for each step, the comment should be in present tense without any additional explanation.
