export interface Person {
  id: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  type: 'lead' | 'contact' | 'customer';
  company_id?: string;
  owner_id?: string;
  score?: number;
  tags?: string[];
  custom_fields?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface Company {
  id: string;
  name: string;
  domain?: string;
  phone?: string;
  address?: string;
  industry?: string;
  size?: 'startup' | 'small' | 'medium' | 'large' | 'enterprise';
  annual_revenue?: number;
  website?: string;
  owner_id?: string;
  tags?: string[];
  created_at: string;
  updated_at: string;
}
