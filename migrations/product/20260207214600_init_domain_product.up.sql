create schema if not exists product;

CREATE TABLE product.categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES product.categories(id) ON DELETE SET NULL,
    name JSONB NOT NULL,  -- Multi-language: {"en-US": "Water Sports", "id-ID": "Olahraga Air"}
    slug JSONB NOT NULL,  -- Multi-language: {"en-US": "water-sports", "id-ID": "olahraga-air"}
    description JSONB,  -- Multi-language descriptions
    icon_url TEXT,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Generated column for default language
    name_default TEXT GENERATED ALWAYS AS (name->>'en-US') STORED,
    slug_default TEXT GENERATED ALWAYS AS (slug->>'en-US') STORED,
    UNIQUE(slug_default)
);

-- Products (tours, activities)
CREATE TABLE product.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL,  -- No FK, loose coupling to merchant.merchants
    category_id UUID REFERENCES product.categories(id) ON DELETE SET NULL,

    -- Multi-language fields (JSONB format: {"en-US": "value", "id-ID": "value"})
    name JSONB NOT NULL,  -- {"en-US": "Bali Beach Tour", "id-ID": "Tur Pantai Bali"}
    slug JSONB NOT NULL,  -- {"en-US": "bali-beach-tour", "id-ID": "tur-pantai-bali"}
    description JSONB NOT NULL,  -- {"en-US": "Experience...", "id-ID": "Rasakan..."}
    highlights JSONB,  -- {"en-US": ["Point 1", "Point 2"], "id-ID": ["Poin 1", "Poin 2"]}
    what_included JSONB,  -- Same structure as highlights
    what_excluded JSONB,  -- Same structure as highlights

    -- Generated column for default language (performance optimization)
    name_default TEXT GENERATED ALWAYS AS (name->>'en-US') STORED,
    slug_default TEXT GENERATED ALWAYS AS (slug->>'en-US') STORED,

    duration_minutes INT,  -- NULL if variable
    min_participants INT DEFAULT 1,
    max_participants INT,
    difficulty_level VARCHAR(20),  -- easy, moderate, hard
    age_restriction INT,  -- Minimum age, NULL if no restriction
    status VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft, published, archived
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(merchant_id, slug_default)
);

-- Multi-language validation function (ensures BCP 47 format and required default language)
CREATE OR REPLACE FUNCTION validate_product_translations()
RETURNS TRIGGER AS $$
BEGIN
  -- Ensure 'en-US' (default language) exists
  IF NOT (NEW.name ? 'en-US') THEN
    RAISE EXCEPTION 'Default language (en-US) is required for name';
END IF;

  IF NOT (NEW.description ? 'en-US') THEN
    RAISE EXCEPTION 'Default language (en-US) is required for description';
END IF;

  IF NOT (NEW.slug ? 'en-US') THEN
    RAISE EXCEPTION 'Default language (en-US) is required for slug';
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_product_translations
    BEFORE INSERT OR UPDATE ON product.products
    FOR EACH ROW
    EXECUTE FUNCTION validate_product_translations();

-- Product images
CREATE TABLE product.product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES product.products(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    alt_text JSONB,  -- Multi-language: {"en-US": "Beach sunset", "id-ID": "Matahari terbenam pantai"}
    is_primary BOOLEAN DEFAULT false,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Product pricing (can have multiple tiers: adult, child, senior)
CREATE TABLE product.product_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES product.products(id) ON DELETE CASCADE,
    price_type VARCHAR(50) NOT NULL,  -- adult, child, senior, group
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    min_quantity INT DEFAULT 1,
    max_quantity INT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(product_id, price_type)
);

-- Product locations (meeting point, pickup points)
CREATE TABLE product.product_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES product.products(id) ON DELETE CASCADE,
    location_type VARCHAR(50) NOT NULL,  -- meeting_point, pickup_point, drop_off_point
    name VARCHAR(255) NOT NULL,
    address JSONB NOT NULL,  -- {street, city, state, country, postal_code}
    coordinates POINT,  -- PostGIS point (lat, lng)
    instructions TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_merchant ON product.products(merchant_id);
CREATE INDEX idx_products_category ON product.products(category_id);
CREATE INDEX idx_products_status ON product.products(status);
CREATE INDEX idx_products_name_default ON product.products(name_default);
CREATE INDEX idx_products_slug_default ON product.products(slug_default);

-- GIN indexes for JSONB multi-language fields (enables fast language-specific queries)
CREATE INDEX idx_products_name_jsonb ON product.products USING GIN (name);
CREATE INDEX idx_products_description_jsonb ON product.products USING GIN (description);

-- Expression indexes for commonly queried languages
CREATE INDEX idx_products_name_en ON product.products ((name->>'en-US'));
CREATE INDEX idx_products_name_id ON product.products ((name->>'id-ID'));

CREATE INDEX idx_product_images_product ON product.product_images(product_id);