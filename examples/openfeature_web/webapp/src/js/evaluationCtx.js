import { v4 as uuidv4 } from 'uuid';

export const defaultAnonymousCtx = {
    targetingKey: uuidv4(),
    anonymous: true,
};

export const defaultAuthenticatedCtx = {
    targetingKey: uuidv4(),
    anonymous: false,
    firstname: 'John',
    lastname: 'Doe',
    email: 'contact@gofeatureflag.org',
    company: 'GO Feature Flag',
    companyId: 1,
    beta: true,
};