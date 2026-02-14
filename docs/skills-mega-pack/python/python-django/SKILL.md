---
name: python-django
description: Django development with REST framework, ORM optimization, async views, and production deployment
---

# Django Development

## Project Setup
- Use Django 5.x with async support
- Organize with apps: each app owns one bounded context
- Use `django-environ` for configuration from environment
- Use `django-extensions` for development productivity

## Models
```python
class User(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    email = models.EmailField(unique=True, db_index=True)
    name = models.CharField(max_length=255)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        ordering = ["-created_at"]
        indexes = [models.Index(fields=["email", "created_at"])]
```

## Django REST Framework
- Use `ModelSerializer` for CRUD, custom serializers for complex logic
- Use `ViewSet` + `Router` for standard CRUD endpoints
- Use `permission_classes` for authorization
- Use `FilterSet` with `django-filter` for query filtering
- Implement pagination: `CursorPagination` for large datasets

## ORM Optimization
- Use `select_related` for ForeignKey (JOIN)
- Use `prefetch_related` for ManyToMany and reverse FK
- Use `only()`/`defer()` to limit selected columns
- Use `Subquery` and `OuterRef` for complex queries
- Use `bulk_create`/`bulk_update` for batch operations
- Monitor queries with `django-debug-toolbar`

## Testing
```python
class TestUserAPI(APITestCase):
    def setUp(self):
        self.user = UserFactory()
        self.client.force_authenticate(self.user)

    def test_list_users(self):
        response = self.client.get("/api/v1/users/")
        assert response.status_code == 200
        assert len(response.data["results"]) >= 1
```

## Production
- Use Gunicorn with Uvicorn workers for async support
- Configure `ALLOWED_HOSTS`, `SECURE_SSL_REDIRECT`, `CSRF_TRUSTED_ORIGINS`
- Use WhiteNoise for static files or serve from CDN
- Set up proper logging configuration
- Use database connection pooling with `django-db-connection-pool`
