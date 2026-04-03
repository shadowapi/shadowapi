import fs from 'fs'
import path from 'path'

import { FlatCompat } from '@eslint/eslintrc'
import js from '@eslint/js'
import typescriptPlugin from '@typescript-eslint/eslint-plugin'
import typescriptParser from '@typescript-eslint/parser'
import importPlugin from 'eslint-plugin-import'
import prettier from 'eslint-plugin-prettier'
import promise from 'eslint-plugin-promise'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import simpleImportSort from 'eslint-plugin-simple-import-sort'
import unusedImports from 'eslint-plugin-unused-imports'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const compat = new FlatCompat({ baseDirectory: __dirname })

// read tsconfig paths -> core modules
function readTSPathAliases(tsconfigFile) {
  try {
    const raw = fs.readFileSync(path.resolve(__dirname, tsconfigFile), 'utf8')
    const json = JSON.parse(raw)
    const paths = json?.compilerOptions?.paths || {}

    return Object.keys(paths)
      .map((k) => k.replace(/\/\*$/, ''))
      .filter(Boolean)
  } catch {
    return []
  }
}
const coreModulesFromTSPaths = readTSPathAliases('./tsconfig.json')

export default [
  {
    ignores: ['**/dist', '**/build', '**/node_modules', '**/coverage', '**/public'],
  },

  js.configs.recommended,

  // TypeScript recommended config
  ...compat.extends('plugin:@typescript-eslint/recommended', 'plugin:import/typescript'),

  // React plugin recommended
  ...compat.extends('plugin:react/recommended'),
  // Promise plugin
  ...compat.extends('plugin:promise/recommended'),
  // Prettier MUST be last to override formatting rules
  ...compat.extends('plugin:prettier/recommended'),

  {
    languageOptions: {
      parser: typescriptParser,
      parserOptions: {
        ecmaVersion: 2024,
        sourceType: 'module',
        ecmaFeatures: { jsx: true },
      },
      globals: {
        browser: 'readonly',
        node: 'readonly',
        es6: 'readonly',
      },
    },

    plugins: {
      promise,
      prettier,
      react,
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
      'simple-import-sort': simpleImportSort,
      'unused-imports': unusedImports,
      import: importPlugin,
      '@typescript-eslint': typescriptPlugin,
    },

    settings: {
      react: { version: 'detect' },
      'import/resolver': {
        node: { moduleDirectory: ['node_modules', 'src'], extensions: ['.js', '.jsx', '.ts', '.tsx'] },
        typescript: { alwaysTryTypes: true, project: './tsconfig.json' },
      },
      'import/parsers': { '@typescript-eslint/parser': ['.ts', '.tsx'] },
      'import/core-modules': coreModulesFromTSPaths,
    },

    rules: {
      // Prettier
      'prettier/prettier': 'error',

      // Imports and unused
      '@typescript-eslint/no-unused-vars': 'off',
      'unused-imports/no-unused-imports': 'error',
      'unused-imports/no-unused-vars': 'off',

      'import/extensions': ['error', 'never', { ts: 'never', tsx: 'never', js: 'never', jsx: 'never', json: 'always' }],
      'import/no-cycle': 'warn',
      'import/no-dynamic-require': 'off',
      'import/no-extraneous-dependencies': 'off',
      'import/no-named-as-default-member': 'off',
      'import/no-unresolved': ['error', { ignore: ['^@theme', '^@docusaurus', '^@generated'] }],
      'import/prefer-default-export': 'off',
      'import/order': 'off',

      // Sorting (separate groups + alias for "@/")
      'simple-import-sort/imports': [
        'error',
        {
          groups: [
            ['^\\u0000'],
            ['^(assert|buffer|child_process|crypto|events|fs|http|net|os|path|stream|tls|util|zlib)(/.*|$)'],
            ['^react'],
            ['^@?\\w'],
            ['^@(/|$)'],
            ['^(assets|components|config|hooks|plugins|store|styled|themes|utils|contexts)(/.*|$)'],
            ['^\\.\\.(?!/?$)', '^\\.\\./?$'],
            ['^\\./(?=.*/)(?!/?$)', '^\\.(?!/?$)', '^\\./?$'],
          ],
        },
      ],
      'simple-import-sort/exports': 'error',

      // JS
      'no-console': 'off',
      'no-debugger': 'error',
      'no-alert': 'warn',
      'no-await-in-loop': 'off',
      'no-return-assign': 'warn',
      'no-unused-expressions': 'off',
      '@typescript-eslint/no-unused-expressions': ['error', { allowShortCircuit: true, allowTernary: true, allowTaggedTemplates: true }],
      'no-extra-parens': 'off',
      'global-require': 'off',

      // React
      'react/jsx-filename-extension': ['error', { extensions: ['.jsx', '.tsx'] }],
      'react/jsx-props-no-spreading': 'off',
      'react/prop-types': 'off',
      'react/require-default-props': 'off',
      'react/no-array-index-key': 'warn',
      'react/no-unused-prop-types': 'off',
      'react/button-has-type': ['error', { reset: true }],
      'react/jsx-uses-react': 'off',
      'react/react-in-jsx-scope': 'off',
      'react/no-unescaped-entities': 'off',
      'react/destructuring-assignment': 'off',
      'react/jsx-pascal-case': 'error',
      'react/jsx-key': 'off',
      'react/no-unstable-nested-components': 'warn',
      'react/jsx-no-useless-fragment': ['warn', { allowExpressions: true }],
      'react/function-component-definition': 'off',
      'react/jsx-no-constructed-context-values': 'warn',
      'react/jsx-no-leaked-render': ['warn', { validStrategies: ['coerce'] }],

      // Hooks
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': ['warn', { enableDangerousAutofixThisMayCauseInfiniteLoops: true }],

      // HMR safety
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],

      // Style
      'linebreak-style': 'off',
      'lines-between-class-members': ['error', 'always', { exceptAfterSingleLine: true }],
      'padding-line-between-statements': [
        'error',
        { blankLine: 'always', prev: '*', next: 'return' },
        { blankLine: 'always', prev: ['const', 'let', 'var'], next: '*' },
        { blankLine: 'any', prev: ['const', 'let', 'var'], next: ['const', 'let', 'var'] },
      ],

      // Control flow
      'no-continue': 'off',
      'no-multi-assign': 'off',
      'no-empty': 'warn',
      'no-nested-ternary': 'off',
      'no-new': 'off',
      'no-param-reassign': 'off',
      'no-plusplus': 'off',
      'no-prototype-builtins': 'off',
      'no-restricted-syntax': [
        'error',
        { selector: 'ForInStatement', message: 'Use Object.{keys,values,entries}, and iterate over the resulting array.' },
        { selector: 'LabeledStatement', message: 'Labels are a form of GOTO; using them makes code confusing.' },
        { selector: 'WithStatement', message: '`with` is disallowed in strict mode.' },
      ],
      'no-underscore-dangle': 'off',
      'prefer-promise-reject-errors': 'off',
      'promise/catch-or-return': 'warn',
      'promise/always-return': 'off',
      'promise/no-callback-in-promise': 'off',

      // TS
      '@typescript-eslint/explicit-module-boundary-types': 'off',
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-empty-interface': 'off',
      '@typescript-eslint/no-empty-function': 'off',
      '@typescript-eslint/ban-ts-comment': 'warn',
      '@typescript-eslint/ban-types': 'off',
      '@typescript-eslint/no-non-null-assertion': 'off',
      '@typescript-eslint/no-use-before-define': ['error'],
      '@typescript-eslint/no-shadow': ['error'],
      '@typescript-eslint/no-useless-constructor': 'error',
      '@typescript-eslint/no-namespace': 'off',
      '@typescript-eslint/no-var-requires': 'off',

      // Vars (JS variants off; TS handles them)
      'no-shadow': 'off',
      'no-undef': 'off',
      'no-unexpected-multiline': 'off',
      'no-unused-vars': 'off',
      'no-useless-constructor': 'off',
      'no-use-before-define': 'off',

      // Misc
      'max-classes-per-file': 'off',
      'default-case': 'off',
      'no-bitwise': 'off',
      camelcase: 'off',
      'no-constant-condition': 'warn',
      'prefer-template': 'error',
      'prefer-const': 'error',
      'no-var': 'error',
      'object-shorthand': 'error',
      'prefer-destructuring': ['error', { object: true, array: false }],
    },
  },

  { files: ['**/*.ts', '**/*.tsx'], rules: {} },

  { files: ['**/*.js', '**/*.jsx'], rules: { '@typescript-eslint/no-var-requires': 'off' } },

  { files: ['*.config.js', '*.config.ts'], rules: { 'import/no-extraneous-dependencies': 'off' } },
]
