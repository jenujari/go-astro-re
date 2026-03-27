DELETE FROM rule_versions WHERE rule_id IN (
    'moon_in_rohini',
    'saturn_aspect_on_moon',
    'mars_in_own_sign',
    'jupiter_exalted',
    'sun_debilitated'
);

DELETE FROM rules WHERE id IN (
    'moon_in_rohini',
    'saturn_aspect_on_moon',
    'mars_in_own_sign',
    'jupiter_exalted',
    'sun_debilitated'
);
