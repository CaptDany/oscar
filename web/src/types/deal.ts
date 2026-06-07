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

export interface DealLineItem {
  id: string;
  deal_id: string;
  product_id?: string;
  product_name?: string;
  product_sku?: string;
  quantity: number;
  unit_price: number;
  discount_pct: number;
  total: number;
  created_at: string;
  updated_at: string;
}

export interface CreateLineItemRequest {
  product_id?: string;
  quantity: number;
  unit_price: number;
  discount_pct?: number;
}

export interface UpdateLineItemRequest {
  product_id?: string;
  quantity?: number;
  unit_price?: number;
  discount_pct?: number;
}

export interface Stage {
  id: string;
  pipeline_id: string;
  name: string;
  position: number;
  probability: number;
  stage_type?: 'open' | 'won' | 'lost';
}
