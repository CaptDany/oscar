export interface Deal {
  id: string;
  title: string;
  value: number;
  currency?: string;
  status?: 'open' | 'won' | 'lost';
  stage_id?: string;
  pipeline_id?: string;
  company_id?: string;
  person_id?: string;
  owner_id?: string;
  expected_close_date?: string;
  tags?: string[];
  created_at: string;
  updated_at: string;
}

export interface Pipeline {
  id: string;
  name: string;
  currency?: string;
  is_default?: boolean;
  stages?: Stage[];
  created_at: string;
  updated_at: string;
}

export interface Stage {
  id: string;
  pipeline_id: string;
  name: string;
  position: number;
  probability: number;
  stage_type?: 'open' | 'won' | 'lost';
}
