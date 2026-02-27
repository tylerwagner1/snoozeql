---
phase: quick-011
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - web/src/pages/CloudAccountsPage.tsx
autonomous: true
---

<objective>
Show "Coming Soon!" message when user selects GCP provider in Cloud Accounts page.

Purpose: GCP Cloud SQL integration was never fully implemented. Instead of showing a broken form, display a friendly "Coming Soon!" message.
Output: GCP button still visible but selecting it shows info message instead of credentials form.
</objective>

<context>
@web/src/pages/CloudAccountsPage.tsx - Current cloud accounts form with AWS/GCP toggle
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add "Coming Soon" message for GCP provider</name>
  <files>web/src/pages/CloudAccountsPage.tsx</files>
  <action>
When form.provider === 'gcp' (and not editing an existing account), instead of showing the credential fields, show:

1. Keep the provider toggle buttons (AWS/GCP) so user can still switch back
2. When GCP is selected, render a styled info box with:
   - A Rocket or Construction icon (use lucide-react)
   - Heading: "GCP Support Coming Soon!"
   - Text: "We're working on GCP Cloud SQL integration. AWS RDS support is fully available."
   - A button to switch back to AWS: "Connect AWS Instead"

3. Hide the credential fields (Access Key, Secret, Regions) when GCP is selected
4. Hide the submit button when GCP is selected

Implementation:
- Add conditional rendering after the provider toggle section
- Check: `form.provider === 'gcp' && !editingAccount`
- If true: render Coming Soon message
- If false: render normal form fields
  </action>
  <verify>
1. `cd web && npm run build` passes
2. Manual check: Open Cloud Accounts page, click "Connect Account", click GCP button - should see Coming Soon message
3. Click "Connect AWS Instead" or AWS button - should see normal form
  </verify>
  <done>
- GCP button shows Coming Soon message
- AWS form continues to work normally
- Build passes
  </done>
</task>

</tasks>

<verification>
- `cd web && npm run build` - TypeScript compiles without errors
- Visual check: GCP shows Coming Soon, AWS shows credentials form
</verification>

<success_criteria>
- GCP selection shows friendly "Coming Soon!" message
- AWS credential form still works
- User can switch between AWS and GCP to see the difference
- Build passes without errors
</success_criteria>

<output>
After completion, create `.planning/quick/011-gcp-cloud-account-coming-soon/011-SUMMARY.md`
</output>
