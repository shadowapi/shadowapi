"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const test_1 = require("@playwright/test");
(0, test_1.test)('Signup page is accessible to guests and shows form', async ({ page }) => {
    await page.goto('http://localtest.me/signup');
    await (0, test_1.expect)(page.getByLabel('First Name')).toBeVisible();
    await (0, test_1.expect)(page.getByRole('button', { name: /^Sign Up$/i })).toBeVisible();
});
//# sourceMappingURL=signup.test.js.map