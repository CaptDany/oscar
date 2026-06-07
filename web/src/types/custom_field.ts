export type FieldType = 'text' | 'number' | 'date' | 'select' | 'multi_select' | 'boolean' | 'url' | 'currency';
export type EntityType = 'person' | 'company' | 'deal';

export interface CustomFieldDefinition {
  id: string;
  tenant_id: string;
  entity_type: EntityType;
  field_key: string;
  label: string;
  field_type: FieldType;
  options?: any;
  is_required: boolean;
  show_in_list: boolean;
  show_in_card: boolean;
  position: number;
  role_visibility?: string[];
  created_at: string;
  updated_at: string;
}
