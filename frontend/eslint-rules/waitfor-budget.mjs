/** vitest kills a test at this many ms unless it declares its own budget. */
const VITEST_DEFAULT_TIMEOUT = 5000;

const TEST_NAMES = new Set(['it', 'test']);
const MODIFIERS = new Set(['only', 'skip', 'concurrent', 'sequential', 'fails', 'todo']);

const isTestIdentifier = (node) => node?.type === 'Identifier' && TEST_NAMES.has(node.name);

const isMemberOfTest = (node, allowed) =>
    node?.type === 'MemberExpression' &&
    isTestIdentifier(node.object) &&
    node.property?.type === 'Identifier' &&
    allowed.has(node.property.name);

const isTestCall = (node) => {
    if (node?.type !== 'CallExpression') {
        return false;
    }

    const { callee } = node;

    return (
        isTestIdentifier(callee) ||
        isMemberOfTest(callee, MODIFIERS) ||
        (callee.type === 'CallExpression' && isMemberOfTest(callee.callee, new Set(['each']))) ||
        (callee.type === 'TaggedTemplateExpression' && isMemberOfTest(callee.tag, new Set(['each'])))
    );
};

const numberOf = (node) => (node?.type === 'Literal' && typeof node.value === 'number' ? node.value : null);

const timeoutIn = (node) => {
    if (node?.type !== 'ObjectExpression') {
        return null;
    }

    const property = node.properties.find(
        (candidate) =>
            candidate.type === 'Property' &&
            !candidate.computed &&
            ((candidate.key.type === 'Identifier' && candidate.key.name === 'timeout') ||
                (candidate.key.type === 'Literal' && candidate.key.value === 'timeout')),
    );

    return property ? numberOf(property.value) : null;
};

const declaredBudget = (testCall) => {
    const [, second, third] = testCall.arguments;

    return timeoutIn(second) ?? numberOf(third) ?? VITEST_DEFAULT_TIMEOUT;
};

const waitForBudget = (node) => {
    const { callee } = node;
    const isWaitFor =
        (callee.type === 'Identifier' && callee.name === 'waitFor') ||
        (callee.type === 'MemberExpression' &&
            callee.object?.type === 'Identifier' &&
            callee.object.name === 'vi' &&
            callee.property?.type === 'Identifier' &&
            callee.property.name === 'waitFor');

    return isWaitFor ? timeoutIn(node.arguments[1]) : null;
};

/** @type {import('eslint').Rule.RuleModule} */
export default {
    create(context) {
        return {
            CallExpression(node) {
                const budget = waitForBudget(node);
                if (budget === null) {
                    return;
                }

                let enclosing = node.parent;
                while (enclosing && !isTestCall(enclosing)) {
                    enclosing = enclosing.parent;
                }

                if (!enclosing) {
                    return;
                }

                const declared = declaredBudget(enclosing);
                if (budget < declared) {
                    return;
                }

                context.report({ data: { budget, declared }, messageId: 'unreachable', node });
            },
        };
    },
    meta: {
        docs: {
            description: 'A waitFor budget must be reachable within the timeout its test declares.',
        },
        messages: {
            unreachable:
                'This waitFor waits up to {{budget}}ms inside a test vitest kills at {{declared}}ms, so its error and DOM dump are discarded. Raise the it()/test() timeout above {{budget}}ms.',
        },
        schema: [],
        type: 'problem',
    },
};
