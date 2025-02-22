import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import importPlugin from "eslint-plugin-import";

// @reactima added
import prettier from 'eslint-plugin-prettier'
import promise from 'eslint-plugin-promise'
import react from 'eslint-plugin-react'
import typescriptPlugin from '@typescript-eslint/eslint-plugin'
import simpleImportSort from 'eslint-plugin-simple-import-sort'

export default tseslint.config(
  { ignores: ['dist'] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      promise,
      prettier,
      react,
      'react-hooks': reactHooks,
        'simple-import-sort': simpleImportSort,
      'react-refresh': reactRefresh,
        import: importPlugin,
      '@typescript-eslint': typescriptPlugin,
    },

    settings: {
      react: { version: 'detect' },
      'import/resolver': {
        node: {
          moduleDirectory: ['node_modules'],
        },
        typescript: { alwaysTryTypes: true },
      },
    },


    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
        // @reactima added
        'no-extra-parens': 'warn',
        'prettier/prettier': [
            'warn',
            {
                trailingComma: 'es5',
                semi: false,
                singleQuote: true,
                printWidth: 120,
            },
        ],
        'class-methods-use-this': 'off',
        'consistent-return': 'off',
      'import/extensions': ['error', 'never', { ts: 'never', tsx: 'never', js: 'never', jsx: 'never', json: 'always' }],
        'import/no-cycle': 'off',
        'import/no-dynamic-require': 'off',
        'import/no-extraneous-dependencies': 'off',
        'import/no-named-as-default-member': 'off',
        'import/no-unresolved': ['error', { ignore: ['^@theme', '^@docusaurus', '^@generated'] }],
        'import/prefer-default-export': 'off',
        'import/order': ['off', { 'newlines-between': 'always' }],
        'simple-import-sort/imports': [
            'error',
            {
                groups: [
                    [
                        '^\\u0000',
                        '^(assert|buffer|child_process|crypto|events|fs|http|net|os|path|stream|tls|util|zlib)(/.*|$)',
                        '^react',
                        '^@?\\w',
                        '^(assets|components|config|hooks|plugins|store|styled|themes|utils|contexts)(/.*|$)',
                        '^\\.\\.(?!/?$)',
                        '^\\.\\./?$',
                        '^\\./(?=.*/)(?!/?$)',
                        '^\\.(?!/?$)',
                        '^\\./?$',
                    ],
                ],
            },
        ],

        'global-require': 'off',
        'jsx-a11y/anchor-is-valid': 'off',
        'jsx-a11y/click-events-have-key-events': 'off',
        'jsx-a11y/label-has-associated-control': 'off',
        'jsx-a11y/label-has-for': 'off',
        'jsx-a11y/no-autofocus': 'off',
        'jsx-a11y/no-static-element-interactions': 'off',
        'linebreak-style': 'off',
        'lines-between-class-members': ['error', 'always', { exceptAfterSingleLine: true }],
        'no-alert': 'off',
        'no-continue': 'off',
        'no-multi-assign': 'off',
        'no-await-in-loop': 'off',
        'no-empty': 'off',
        'no-console': ['warn', { allow: ['info', 'warn', 'error'] }],
        'no-nested-ternary': 'off',
        'no-new': 'off',
        'no-param-reassign': 'off',
        'no-plusplus': 'off',
        'no-prototype-builtins': 'off',
        'no-restricted-syntax': [
            'error',
            {
                selector: 'ForInStatement',
                message: 'Use Object.{keys,values,entries}, and iterate over the resulting array.',
            },
            { selector: 'LabeledStatement', message: 'Labels are a form of GOTO; using them makes code confusing.' },
            { selector: 'WithStatement', message: '`with` is disallowed in strict mode.' },
        ],
        'no-return-assign': 'off',
        'no-underscore-dangle': 'off',
        'no-unused-expressions': 'off',
        'prefer-promise-reject-errors': 'off',
        'promise/catch-or-return': 'off',
        'promise/always-return': 'off',
        'promise/no-callback-in-promise': 'off',
        'react/button-has-type': ['error', { reset: true }],
        'react/jsx-filename-extension': ['error', { extensions: ['.js', '.jsx', '.ts', '.tsx', 'mdx'] }],
        'react/jsx-props-no-spreading': 'off',
        'react/no-array-index-key': 'off',
        'react/no-unused-prop-types': 'off',
        'react/prop-types': 'off',
        'react/require-default-props': 'off',
        'react/no-unused-expressions': 'off',
        'react-hooks/exhaustive-deps': ['warn', { enableDangerousAutofixThisMayCauseInfiniteLoops: true }],
        'react-hooks/rules-of-hooks': 'error',
        '@typescript-eslint/no-unused-expressions': 'off',
        '@typescript-eslint/ban-ts-ignore': 'off',
        '@typescript-eslint/ban-types': 'off',
        '@typescript-eslint/camelcase': 'off',
        '@typescript-eslint/explicit-module-boundary-types': 'off',
        '@typescript-eslint/explicit-function-return-type': 'off',
        '@typescript-eslint/interface-name-prefix': 'off',
        '@typescript-eslint/no-empty-interface': 'off',
        '@typescript-eslint/no-empty-function': 'off',
        '@typescript-eslint/no-explicit-any': 'off',
        '@typescript-eslint/no-namespace': 'off',
        '@typescript-eslint/no-shadow': ['error'],
        '@typescript-eslint/no-unused-vars': 'off',
        '@typescript-eslint/no-useless-constructor': 'error',
        '@typescript-eslint/no-var-requires': 'off',
        'babel/no-unused-expressions': 'off',
        'no-shadow': 'off',
        'no-undef': 'off',
        'no-unexpected-multiline': 'off',
        'no-unused-vars': 'off',
        'no-useless-constructor': 'off',
        'no-use-before-define': 'off',
        '@typescript-eslint/no-use-before-define': ['error'],
        '@typescript-eslint/no-non-null-assertion': 'off',
        '@typescript-eslint/ban-ts-comment': 'off',
        'max-classes-per-file': 'off',
        'default-case': 'off',
        'no-bitwise': 'off',
        'react/jsx-uses-react': 'off',
        'react/no-unescaped-entities': 'off',
        'react/react-in-jsx-scope': 'off',
        camelcase: 'off',
        'no-constant-condition': 'off',

    },
  },
)
